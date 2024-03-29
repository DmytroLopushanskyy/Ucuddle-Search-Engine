import math
import os
import time

import elasticsearch

import config
import logging
import jsonpickle

from datetime import datetime
from elasticsearch import Elasticsearch


class TaskManager:
    def __init__(self):
        self.es_client = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                                       http_auth=(os.environ['Username'], os.environ['Password']))
        # self.es_client = Elasticsearch()

        index_elastic_links = os.environ["INDEXES_ELASTIC_LINKS"]
        self.index_elastic_links = index_elastic_links.split()[config.POS_ELASTIC_INDEX_LINKS]

        self.indexes_elastic_sites = dict()
        self.indexes_elastic_sites["ukr"] = os.environ["INDEX_ELASTIC_UKR_COLLECTED_DATA"]
        self.indexes_elastic_sites["ru"] = os.environ["INDEX_ELASTIC_RU_COLLECTED_DATA"]

        self.last_id_in_index_sites = -1

        for index_elastic_sites in self.indexes_elastic_sites.values():
            try:
                res = self.es_client.get(index=index_elastic_sites, id=0)
                print("res -- ", res)
                index_exist = True
            except elasticsearch.exceptions.NotFoundError:
                index_exist = False

            if index_exist:
                last_id_in_index_sites = self.get_last_site_id_in_index(index_elastic_sites)
                self.last_id_in_index_sites = max(self.last_id_in_index_sites, last_id_in_index_sites)

    def create_new_index(self, index_name):
        logging.debug("Creating new index")
        if index_name != "la_links_to_parse1":
            res = self.es_client.indices.create(index=index_name, ignore=400)

            if res.get('status', 0) != 0:
                print("Error reason: ", res['error']['reason'])
                return res['status']

            logging.debug("Elastic response for creating new index: ", res)
        return 200

    def retrieve_links(self, num_links, taken_value):
        query = {
            "query": {
                "bool": {
                    "must": [
                        {"term": {"taken": taken_value}},
                        {"term": {"parsed": False}}
                    ]
                }
            },
            "size": num_links,

            "sort": {
                "link_id": {
                    "order": "asc",
                },
            }
        }

        try:
            res = self.es_client.search(
                index=self.index_elastic_links,
                body=jsonpickle.encode(query, unpicklable=False),
                request_timeout=150,
            )

            links = dict()
            links["links"] = []
            missed_hits = 0
            for hit in res["hits"]["hits"]:
                if self.set_parsed_link(hit["_id"], 0, True) == 0:
                    links["links"].append([hit["_source"]["link"], hit["_id"]])
                else:
                    missed_hits += 1

            return links, missed_hits

        except Exception as e:
            print("retrieve_links(): search response is empty to ", self.index_elastic_links)
            print("retrieve_links(): error --  ", e)
            return {"links": []}, -1

    def add_new_links(self, links_index_name, links_lst):
        last_link_id = self.get_last_link_id()

        print("last_link_id before adding links -- ", last_link_id)
        for link in links_lst:
            res = self.es_client.index(
                index=links_index_name,
                id=last_link_id,
                body={
                    "link_id": last_link_id,
                    "link": link,
                    "taken": False,
                    "parsed": False,
                    "added_at_time": datetime.now()
                },
                request_timeout=150
            )

            last_link_id += 1
            if last_link_id % 100 == 0:
                print("Currently added last_link_id -- ", last_link_id)

        print("last_link_id after adding links -- ", last_link_id)

        self.update_last_link_id(last_link_id)
        print("All links were successfully added !!!")

    def get_last_link_id(self):
        index_name = os.environ['INDEX_CONFIG_ELASTIC']

        res = self.es_client.get(index=index_name, id=0)
        print("get_last_link_id -- ", res)

        return res["_source"][os.environ["LAST_LINK_ID_KEY_ELASTIC"]]

    def update_last_link_id(self, last_id):
        body_dict = {
            "doc": {
                os.environ["LAST_LINK_ID_KEY_ELASTIC"]: last_id,
            }
        }

        res = self.es_client.update(index=os.environ['INDEX_CONFIG_ELASTIC'],
                                    id=0,
                                    body=body_dict)

        print(res['result'])

    def set_parsed_link(self, parsed_link_id, body_key, body_value):
        if body_key == 0:
            body_key = "taken"

        elif body_key == 1:
            body_key = "parsed"

        # it wan not unnecessary to make retry here
        # as crawler make retry by itself
        try:
            res = self.es_client.update(
                index=self.index_elastic_links,
                id=parsed_link_id,
                body={
                    "doc": {
                        body_key: body_value
                    }
                },
                request_timeout=150
            )

            print("set_parsed_link() res -- ", res)

            if res['result'] == 'updated':
                return 0
            return -1
        except Exception as e:
            print("Exception in set_parsed_link() -- ", e)
            return -1

    def get_last_site_id_in_index(self, index_elastic_sites):
        query = {
            "query": {
                "match_all": {},
            },

            "size": 1,

            "sort": {
                "site_id": {
                    "order": "desc",
                },
            },
        }

        waiting_response_time = 0

        # retry in elastic
        for i in range(3):
            time.sleep(waiting_response_time)

            res = self.es_client.search(
                index=index_elastic_sites,
                body=jsonpickle.encode(query, unpicklable=False),
                request_timeout=150
            )

            try:
                return res["hits"]["hits"][0]["_source"]["site_id"] + 1
            except Exception as e:
                logging.error("set_parsed_link() es_client.update invalid response", res)
                logging.error("set_parsed_link() error -- ", e)

            waiting_response_time = math.exp(i + 1)

    def get_last_site_id_in_all_indexes(self):
        last_site_id = -1
        for index_name in self.indexes_elastic_sites.values():
            last_site_id = max(last_site_id, self.get_last_site_id_in_index(index_name))

        return last_site_id
