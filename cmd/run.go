// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	db "github.com/eaciit/dbox"
	_ "github.com/eaciit/dbox/dbc/mongo"
	"github.com/eaciit/toolkit"
	tk "github.com/eaciit/toolkit"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Copy Solr collection to Solr/mongodb",
	Long:  `Copy Solr collection to Solr/mongodb`,
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().IntP("max", "m", -1, "Maximum data to be copied, -1/0 for all")

	runCmd.Flags().StringP("source", "s", "", "Source Collection")
	runCmd.Flags().StringP("target", "t", "", "Target Collection")

	runCmd.Flags().String("source-host", "http://127.0.0.1:8983/solr/", "Solr Source Full URL (with /solr)")
	runCmd.Flags().StringP("source-query", "q", "*:*&sort=id+desc", "Solr Source Query")
	runCmd.Flags().Int("source-rows", 10000, "Solr Source Rows Fetch each Query")
	runCmd.Flags().String("source-cursor", "*", "Solr Source Cursor")

	runCmd.Flags().String("target-host", "http://127.0.0.1:8983/solr/", "Mongo (Mongo) | Solr Source Full URL (with /solr)")
	runCmd.Flags().String("target-db", "", "Database (Mongo)")
	runCmd.Flags().String("target-user", "", "Username Database (Mongo)")
	runCmd.Flags().String("target-pass", "", "Password Database (Mongo)")

	runCmd.Flags().Bool("target-commit", true, "Commit after Post (Solr)")

	runCmd.Flags().String("target-type", "solr", "Target Collection Type : mongo / solr")

}

func run(cmd *cobra.Command, args []string) {
	log.Info("==== Solr Copy Collection ====")

	source := cmd.Flags().Lookup("source").Value.String()
	target := cmd.Flags().Lookup("target").Value.String()

	smax := cmd.Flags().Lookup("max").Value.String()

	sourceHost := cmd.Flags().Lookup("source-host").Value.String()
	sourceQuery := cmd.Flags().Lookup("source-query").Value.String()
	ssourceRows := cmd.Flags().Lookup("source-rows").Value.String()
	sourceCursorMark := cmd.Flags().Lookup("source-cursor").Value.String()

	targetType := cmd.Flags().Lookup("target-type").Value.String()

	targetHost := cmd.Flags().Lookup("target-host").Value.String()
	targetDB := cmd.Flags().Lookup("target-db").Value.String()
	targetUser := cmd.Flags().Lookup("target-user").Value.String()
	targetPass := cmd.Flags().Lookup("target-pass").Value.String()

	targetCommit := cmd.Flags().Lookup("target-commit").Value.String()

	sourceRows, e := strconv.Atoi(ssourceRows)
	if e != nil {
		sourceRows = 10000
	}

	max, e := strconv.Atoi(smax)
	if e != nil {
		max = -1
	}

	// solr url
	targetSolrUrlPost := fmt.Sprintf("%s%s/update/json/docs", targetHost, target)
	targetSolrUrlCommit := fmt.Sprintf("%s%s/update?commit=true", targetHost, target)

	if toolkit.IsNilOrEmpty(source) || toolkit.IsNilOrEmpty(target) {
		log.Fatal("--source and --target is required")
	}

	if targetType == "mongo" {
		if toolkit.IsNilOrEmpty(targetHost) {
			log.Fatal("--target-db is required when --target-type=mongo selected")
		}

		log.WithFields(log.Fields{
			"--source":       source,
			"--target":       target,
			"--max":          max,
			"--source-url":   sourceHost,
			"--source-query": sourceQuery,
			"--source-rows":  sourceRows,
			"--target-type":  targetType,
			"--target-host":  targetHost,
			"--target-db":    targetDB,
			"--target-user":  targetUser,
			"--target-pass":  targetPass,
			"sourceRows":     sourceRows,
		}).Info("Flag Registered")

	} else {
		targetType = "solr" // default solr

		log.WithFields(log.Fields{
			"--source":          source,
			"--target":          target,
			"--max":             max,
			"--source-url":      sourceHost,
			"--source-query":    sourceQuery,
			"--source-rows":     sourceRows,
			"--target-type":     targetType,
			"--target-host":     targetHost,
			"--target-commit":   targetCommit,
			"targetSolrUrlPost": targetSolrUrlPost,
			"sourceRows":        sourceRows,
		}).Info("Flag Registered")
	}

	docClean := []byte{}

	log.Info("==== Start Application ====")

	var mgoConn db.IConnection
	if targetType == "mongo" {
		log.Info("Connecting to Mongo")

		var err error
		mgoConn, err = mgoNewConnection(targetHost, targetDB, targetUser, targetPass)
		if err != nil {
			log.Fatalf("Connection to target mongo failed. %v", err.Error())
		}
		defer mgoConn.Close()
		log.Info("Connected!")
	}

	// connection ok!
	totalData := 0

	for i := 0; ; i++ {
		end := func() bool {
			rowToGet := sourceRows

			// fetch row more than max
			if rowToGet > max && max > 0 {
				rowToGet = max
			}

			// remaining
			if max-totalData < sourceRows && max > 0 {
				rowToGet = max - totalData
			}

			client := http.Client{}
			sourceSolrUrl := fmt.Sprintf("%s%s/select?q=%s&rows=%v&wt=json&cursorMark=%s", sourceHost, source, sourceQuery, rowToGet, sourceCursorMark)
			log.Debugf("Getting Data from %v", sourceSolrUrl)
			resp, err := client.Get(sourceSolrUrl)
			if err != nil {
				log.Error(err.Error())
				return true
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				log.Errorf("Error status %v", resp.StatusCode)
				res, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Debug(err.Error())
					return true
				}
				log.Debug(string(res))
				return true
			}
			res, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err.Error())
				return true
			}

			resMap := map[string]interface{}{}
			err = json.Unmarshal(res, &resMap)
			if err != nil {
				log.Errorf("error unmarshal %v", err.Error())
				log.Debug(string(res))
				return true
			}
			docRes := resMap["response"].(map[string]interface{})
			ii := docRes["docs"].([]interface{})

			log.WithFields(log.Fields{
				"length": len(ii),
			}).Debug("Document Received")
			totalData += len(ii)

			// posting data
			if targetType == "mongo" {
				// connection ok!
				// insert 1-1
				for k, idoc := range ii {
					docMap := idoc.(map[string]interface{})
					delete(docMap, "_version_")

					// mongo = copy id to _id
					if targetType == "mongo" {
						docMap["_id"] = docMap["id"]
						delete(docMap, "id")
					}
					ii[k] = docMap

					qInsert := mgoConn.NewQuery().From(target).Save()
					defer qInsert.Close()
					eInsert := qInsert.Exec(tk.M{}.Set("data", idoc))
					if eInsert != nil {
						log.Error(eInsert)
						return true
					}
				}

			} else {
				for k, doc := range ii {
					docMap := doc.(map[string]interface{})
					delete(docMap, "_version_")
					ii[k] = docMap
				}

				docClean, _ = json.Marshal(ii)

				// solr
				resp2, err := client.Post(targetSolrUrlPost, "application/json", bytes.NewBuffer(docClean))
				if err != nil {
					log.Fatal("fail post", err.Error())
					os.Exit(1)
				}

				defer resp2.Body.Close()

				if resp2.StatusCode != 200 {
					log.Debug(resp.StatusCode)
					oo, _ := ioutil.ReadAll(resp2.Body)
					log.Fatal(string(oo))
					os.Exit(1)
				}
			}
			// end posting data

			sourceCursorMark = resMap["nextCursorMark"].(string)

			log.WithFields(log.Fields{
				"cursor":    sourceCursorMark,
				"totalData": totalData,
			}).Debug("Cursor Mark")

			if len(ii) < sourceRows {
				return true
			}

			if totalData >= max && max >= 0 {
				return true
			}

			return false
		}()

		if end {
			break
		}
	}

	if targetType == "solr" {
		// commit
		client := http.Client{}
		resp, err := client.Get(targetSolrUrlCommit)
		log.Debug(targetSolrUrlCommit)
		if err != nil {
			log.Error(err.Error())
			return
		}
		if resp.StatusCode != 200 {
			log.Errorf("Error status %v", resp.StatusCode)
			res, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Debug(err.Error())
				return
			}
			resp.Body.Close()
			log.Debug(string(res))
			return
		}
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			return
		}
		resp.Body.Close()

		resMap := map[string]interface{}{}
		err = json.Unmarshal(res, &resMap)

		log.Debug(resMap)
		log.Info("Commit Target Solr OK")
	}
	log.Info("=== Done === ")
}

func mgoNewConnection(host, database, user, pass string) (db.IConnection, error) {
	connInfo := &db.ConnectionInfo{
		Host:     host,
		Database: database,
		UserName: user,
		Password: pass,
		Settings: tk.M{}.Set("timeout", 30),
	}

	conn, err := db.NewConnection("mongo", connInfo)
	if err != nil {
		return nil, err
	}

	err = conn.Connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}
