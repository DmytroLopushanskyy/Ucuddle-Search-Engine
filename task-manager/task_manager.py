from datetime import datetime
import os

import jsonpickle
from elasticsearch import Elasticsearch
from logger import get_logger


class _TaskManager:
    _instance = None

    def __init__(self):
        self.logger = get_logger()
        self.logger.log("Initialize TaskManager")
        self.es_client = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                            http_auth=(os.environ['USERNAME'], os.environ['PASSWORD']))

        self.logger.log("Connection opened in TaskManager init")
        # self.es_client = Elasticsearch()

    def create_new_index(self, index_name):
        self.logger.log("Creating new index")
        res = self.es_client.indices.create(index=index_name, ignore=400)

        if res.get('status', 0) != 0:
            self.logger.log("Error reason: ", res['error']['reason'])
            return res['status']

        self.logger.log("Elastic response for creating new index ")
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
        self.logger.log("before elastic")
        res = self.es_client.search(
            index=os.environ['INDEX_ELASTIC_LINKS'],
            body=jsonpickle.encode(query, unpicklable=False)
        )
        self.logger.log("elastic res")
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

            self.logger.log("es_client insert link response")
            last_link_id += 1

        self.logger.log("All links were successfully added !!!")


def get_task_manager_singleton():
    if _TaskManager._instance is None:
        _TaskManager._instance = _TaskManager()
    return _TaskManager._instance
