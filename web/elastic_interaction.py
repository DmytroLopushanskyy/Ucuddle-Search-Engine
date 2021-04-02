import json
import logging
import math
import os
import time

import jsonpickle
from elasticsearch import Elasticsearch

es = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                   http_auth=(os.environ['USERNAME'], os.environ['PASSWORD']))


class MyEncoder(json.JSONEncoder):
    def default(self, o):
        return o.__dict__


def elastic_search(search_line):
    query = {
        "query": {
            "bool": {
                "must": {
                    "multi_match": {
                        "query": search_line,
                        "fuzziness": "AUTO",
                        "minimum_should_match": "100%",
                        "operator": "or",
                        "fields": {
                            "title^5",
                            "content",
                        },
                    },
                },
            },
        }
    }
    hits_list = []
    waiting_response_time = 0

    # TODO: change num waiting cycles after if necessary
    for i in range(3):
        time.sleep(waiting_response_time)
        res = es.search(
            index=os.environ['INDEX_ELASTIC_COLLECTED_DATA'],
            body=jsonpickle.encode(query, unpicklable=False)
        )

        if res['timed_out'] != False or res['_shards']['failed'] != 0 or \
                res.get('status', 0) != 0:
            print("elastic_search(): response error from Elasticsearch -- ")
            print("waiting_response_time -- ", waiting_response_time)
            print("res['_shards']['failed'] -- ", res['_shards']['failed'])
            print("res['timed_out'] -- ", res['timed_out'])
            print("res.get('status', 0) -- ", res.get('status', 0))

        else:
            print("Got %d Hits:" % res['hits']['total']['value'])
            for hit in res['hits']['hits']:
                print('hit["_source"]', hit["_source"]["title"])
                hits_list.append(hit["_source"])

            break

        waiting_response_time = math.exp(i + 1)

    return hits_list
