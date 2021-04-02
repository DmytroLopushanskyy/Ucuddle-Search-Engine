import math
import time
from datetime import datetime
import os
import logging
from typing import List, Any

import jsonpickle
from elasticsearch import Elasticsearch


class TaskManager:
    def __init__(self):
        print("os.environ['ELASTICSEARCH_URL']]", os.environ['ELASTICSEARCH_URL'])

        self.es_client = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                                        http_auth=(os.environ['USERNAME'], os.environ['PASSWORD']))

        # self.es_client = Elasticsearch()

    def create_new_index(self, index_name):
        logging.debug("Creating new index")
        res = self.es_client.indices.create(index=index_name, ignore=400)

        if res.get('status', 0) != 0:
            print("Error reason: ", res['error']['reason'])
            return res['status']

        logging.debug("Elastic response for creating new index: ", res)
        return 200

    def retrieve_links(self, num_links):
        query = {
            "query": {
                "match_all": {}
            },

            "size": num_links,

            "sort": {
                "added_at_time": {
                    "order": "desc",
                }
            }
        }

        waiting_response_time = 0
        links = dict()
        for i in range(5):
            time.sleep(waiting_response_time)
            res = self.es_client.search(
                index=os.environ['INDEX_ELASTIC_LINKS'],
                body=jsonpickle.encode(query, unpicklable=False)
            )

            if res['timed_out'] != False or res['_shards']['failed'] != 0 or \
                    res.get('status', 0) != 0:
                print("retrieve_links(): response error from Elasticsearch -- ")
                print("waiting_response_time -- ", waiting_response_time)
                print("res['_shards']['failed'] -- ", res['_shards']['failed'])
                print("res['timed_out'] -- ", res['timed_out'])
                print("res.get('status', 0) -- ", res.get('status', 0))

            else:
                links["links"] = []
                for hit in res["hits"]["hits"]:
                    links["links"].append(hit["_source"]["link"])

                break

            waiting_response_time = math.exp(i + 1)

        return links

    def add_new_links(self, links_lst, index_name):
        # TODO
        last_link_id = 0
        last_link_id += 1
        missed_links = []
        for link in links_lst:
            waiting_response_time = 0

            for i in range(5):
                time.sleep(waiting_response_time)
                res = self.es_client.index(
                    index=index_name, id=last_link_id,
                    body={"link": link, "added_at_time": datetime.now()}
                )

                if res['result'] != 'created' or res['_shards']['failed'] != 0 or \
                        res.get('status', 0) != 0:
                    print("add_new_links(): response error from Elasticsearch -- ")
                    print("waiting_response_time -- ", waiting_response_time)
                    print("res['_shards']['failed'] -- ", res['_shards']['failed'])
                    print("res['timed_out'] -- ", res['timed_out'])
                    print("res.get('status', 0) -- ", res.get('status', 0))

                    if i == 4:
                        missed_links.append(link)

                else:
                    print("es_client insert link response -- ", res["result"])
                    last_link_id += 1

                    break

                waiting_response_time = math.exp(i + 1)

        if len(missed_links) != 0:
            print("add_new_links() Error: next links was not added")
            for link in missed_links:
                print(link)
            print()
        else:
            print("All links were successfully added !!!")
