import os
import json
import requests
from elasticsearch import Elasticsearch


def update_config_index():
    es_client = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                              http_auth=(os.environ['Username'], os.environ['Password']))

    index_name = os.environ['INDEX_CONFIG_ELASTIC']
    res = es_client.get(index=index_name, id=0)
    print("previous config -- ", res)

    body_dict = {
        "doc": {
            os.environ["LAST_LINK_ID_KEY_ELASTIC"]: 1
        }
    }

    resp = es_client.update(index=os.environ["INDEX_CONFIG_ELASTIC"], id=0,
                            body=body_dict)

    print("update_config_index resp -- ", resp)


if __name__ == '__main__':
    answer = input("Do you want to update LAST_LINK_ID_KEY_ELASTIC -- ")
    if answer.lower() == "yes":
        update_config_index()

    with open(os.path.join("..", "files", "page_rank_test_domains.json"), "r", encoding="utf-8") as f:
        dict_links = json.load(f)

    # slice_id = 0
    # step = 20
    #
    # indexes_names = os.environ["INDEXES_ELASTIC_LINKS"].split()
    # for i in range(2):
    #     url = os.environ["TASK_MANAGER_URL"] + os.environ["TASK_MANAGER_ENDPOINT_ADD_LINKS"]
    #     response = requests.post(url, json={
    #         "links_index_name": indexes_names[i],
    #         "links": dict_links["links"][slice_id: slice_id + step]
    #     })
    #
    #     slice_id += step
    #
    #     print(response, response.reason)

    slice_id = 101
    step = 30

    indexes_names = os.environ["INDEXES_ELASTIC_LINKS"].split()
    url = os.environ["TASK_MANAGER_URL"] + os.environ["TASK_MANAGER_ENDPOINT_ADD_LINKS"]
    response = requests.post(url, json={
        "links_index_name": indexes_names[1],
        "links": dict_links["links"][slice_id: slice_id + step]
    })

    slice_id += step

    print(response, response.reason)
