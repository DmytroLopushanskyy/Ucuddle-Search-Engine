import os
import json
import requests
from elasticsearch import Elasticsearch


def update_config_index():
    es_client = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                               http_auth=(os.environ['USERNAME'], os.environ['PASSWORD']))

    index_name = os.environ['INDEX_CONFIG_ELASTIC']
    res = es_client.get(index=index_name, id=0)
    print("previous config -- ", res)

    # TODO: should be update function
    resp = es_client.index(index=os.environ["INDEX_CONFIG_ELASTIC"], id=0,
                           body={
                               os.environ["LAST_LINK_ID_KEY_ELASTIC"]: 1,
                           })

    print("update_config_index resp -- ", resp)


if __name__ == '__main__':
    update_config_index()

    with open(os.path.join("..", "files", "PG_test2.json"), "r", encoding="utf-8") as f:
        dict_links = json.load(f)

    url = os.environ["TASK_MANAGER_URL"] + os.environ["TASK_MANAGER_ENDPOINT_ADD_LINKS"]
    response = requests.post(url, json=dict_links)

    print(response, response.reason)
