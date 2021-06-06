import json
import math
import os
import time

import jsonpickle
from pprint import pprint
from elasticsearch import Elasticsearch
from bs4 import BeautifulSoup
from flask import Markup

es = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                   http_auth=(os.environ['Username'], os.environ['Password']))


def elastic_search(search_line, lang):
    pre_tag = "*****"
    post_tag = "*-*-*"
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
        },
        "highlight": {
            "pre_tags": [
                pre_tag
            ],
            "post_tags": [
                post_tag
            ],
            "fields": {
                "content": {}
            }
        }
    }

    hits_list = []
    waiting_response_time = 0
    if lang == "uk":
        index_name = os.environ['INDEX_ELASTIC_UKR_COLLECTED_DATA']

    else:
        index_name = os.environ['INDEX_ELASTIC_RU_COLLECTED_DATA']

    # TODO: change num waiting cycles after if necessary
    filter_duplicates = set()
    for i in range(3):
        time.sleep(waiting_response_time)
        res = es.search(
            index=index_name,
            body=jsonpickle.encode(query, unpicklable=False),
            request_timeout=100
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
                # pprint(hit)
                hit["_source"]['highlight'] = hit["highlight"]["content"][0]
                if hit["_source"]["link"] not in filter_duplicates:
                    print('hit["_source"]', hit["_source"]["title"])
                    hit["_source"]["title"] = hit["_source"]["title"].capitalize()

                    # set up bold for highlight
                    soup = BeautifulSoup(hit["_source"]["highlight"])
                    hit["_source"]["highlight"] = soup.get_text()
                    hit["_source"]["highlight"] = hit["_source"]["highlight"].replace(pre_tag, "<b>")
                    hit["_source"]["highlight"] = hit["_source"]["highlight"].replace(post_tag, "</b>")
                    hit["_source"]["highlight"] = Markup(hit["_source"]["highlight"])

                    hits_list.append(hit["_source"])
                    filter_duplicates.add(hit["_source"]["link"])

            break

        waiting_response_time = math.exp(i + 1)

    return hits_list
