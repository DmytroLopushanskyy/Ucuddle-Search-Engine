from datetime import datetime
import os
import logging

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

    # def add_new_data_in_index(self, index_name, data):
    #     self.es_client.index(index=index_name, id=42, body=data)

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
        res = self.es_client.search(
            index=os.environ['INDEX_ELASTIC_LINKS'],
            body=jsonpickle.encode(query, unpicklable=False)
        )

        logging.debug("res", res)
        links = dict()
        links["links"] = []
        for hit in res["hits"]["hits"]:
            links["links"].append(hit["_source"]["link"])

        return links

    def add_new_links(self, links_lst, index_name):
        # TODO
        last_link_id = 0
        last_link_id += 1
        for link in links_lst:
            res = self.es_client.index(
                index=index_name, id=last_link_id,
                body={"link": link, "added_at_time": datetime.now()}
            )

            print("es_client insert link response -- ", res["result"])
            last_link_id += 1

        print("All links were successfully added !!!")
