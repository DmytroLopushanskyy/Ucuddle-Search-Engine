import os
import json
import requests
from elasticsearch import Elasticsearch


def update_config_index():
    es_client = Elasticsearch([os.environ['ELASTICSEARCH_URL']],
                              http_auth=(os.environ['Username'], os.environ['Password']))

    last_link_id = 1
    index_name = os.environ['INDEX_CONFIG_ELASTIC']

    try:
        res = es_client.get(index=index_name, id=0)
        print("previous config -- ", res)

        body_dict = {
            "doc": {
                os.environ["LAST_LINK_ID_KEY_ELASTIC"]: last_link_id
            }
        }

        resp = es_client.update(index=os.environ["INDEX_CONFIG_ELASTIC"], id=0,
                                body=body_dict)

    except:
        print("Create new index for config")
        body_dict = {
            os.environ["LAST_LINK_ID_KEY_ELASTIC"]: last_link_id
        }

        resp = es_client.create(index=os.environ["INDEX_CONFIG_ELASTIC"], id=0,
                                body=body_dict)

    print("update_config_index resp -- ", resp)


if __name__ == '__main__':
    # !!! after changing environment variables reload task_manager
    answer = input("Do you want to update LAST_LINK_ID_KEY_ELASTIC -- ")
    if answer.lower() == "yes":
        update_config_index()

    with open(os.path.join("..", "files", "ua_domains_50.json"), "r", encoding="utf-8") as f:
        dict_links = json.load(f)

    options = ["Create new indexes and fill them with links OR append to the first index_name in INDEXES_ELASTIC_LINKS",
               "Add links to existing index",
               "Reset LAST_LINK_ID_KEY_ELASTIC to 1"]
    print("Choose an option of adding link:")

    for i, option in enumerate(options):
        print(str(i + 1) + ") " + option)

    n_option = int(input())

    if n_option == 1:
        slice_id = 0
        step = 25

        indexes_names = os.environ["INDEXES_ELASTIC_LINKS"].split()
        for i in range(1):
            url = os.environ["TASK_MANAGER_URL"] + os.environ["TASK_MANAGER_ENDPOINT_ADD_LINKS"]
            response = requests.post(url, json={
                "links_index_name": indexes_names[i],
                "links": dict_links["links"][slice_id: slice_id + step]
            })

            slice_id += step

            print(response, response.reason)

    elif n_option == 2:
        slice_id = 0
        step = 30

        indexes_names = os.environ["INDEXES_ELASTIC_LINKS"].split()
        url = os.environ["TASK_MANAGER_URL"] + os.environ["TASK_MANAGER_ENDPOINT_ADD_LINKS"]
        response = requests.post(url, json={
            "links_index_name": indexes_names[1],
            "links": dict_links["links"]
        })

        slice_id += step

        print(response, response.reason)

    elif n_option == 3:
        update_config_index()
