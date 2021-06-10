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


async def elastic_search(search_line, lang):
    pre_tag = "*****"
    post_tag = "*-*-*"
    query = {
        # "profile": True,
        "size": 20,
        "query": {
            "bool": {
                "must": {
                    "multi_match": {
                        "query": search_line,
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
    # index_name = 'production_data1'
    if lang == "uk":
        index_name = os.environ['INDEX_ELASTIC_UKR_COLLECTED_DATA']

    else:
        index_name = os.environ['INDEX_ELASTIC_RU_COLLECTED_DATA']

    # TODO: change num waiting cycles after if necessary
    filter_duplicates = set()
    for i in range(3):
        time.sleep(waiting_response_time)
        try:
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
                    if hit["_source"]["link"] not in filter_duplicates:
                        website = dict()

                        print('hit["_source"]["title"]', hit["_source"]["title"])
                        print('hit["_score"]', hit["_score"])
                        website["title"] = hit["_source"]["title"][0].upper() + hit["_source"]["title"][1:]
                        website["link"] = hit["_source"]["link"]

                        # set up bold for highlight
                        try:
                            hit["_source"]['highlight'] = hit["highlight"]["content"][0]
                            soup = BeautifulSoup(hit["_source"]["highlight"], features="html.parser")
                            hit["_source"]["highlight"] = soup.get_text()
                            hit["_source"]["highlight"] = hit["_source"]["highlight"].replace(pre_tag, "<b>")
                            hit["_source"]["highlight"] = hit["_source"]["highlight"].replace(post_tag, "</b>")
                            website["highlight"] = Markup(hit["_source"]["highlight"])
                        except Exception as e:
                            print("elastic_search(): Error with highlight -- ", e)
                            website["highlight"] = "..."

                        hits_list.append(website)
                        filter_duplicates.add(hit["_source"]["link"])

                # pprint(res)
                break

        except Exception as err:
            print("elastic_search(): error with es.search -- ", err)

        waiting_response_time = math.exp(i + 1)

    return hits_list
