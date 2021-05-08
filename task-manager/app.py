import json
import os
import logging
import config

from flask import make_response, request, abort
from flask import Flask, jsonify

from task_manager import TaskManager

app = Flask(__name__)
app.secret_key = config.FLASK_KEY
app.config['SECRET_KEY'] = config.SECRET_KEY
task_manager = TaskManager()

PACKAGE_SIZE = int(os.environ["NUM_SITES_IN_PACKAGE"])


# TODO: test it
@app.errorhandler(404)
def not_found():
    return make_response(jsonify({'error': 'Not found'}), 404)


def check_elastic_search_status():
    logging.debug("Check ES Status")
    index_elastic_links = os.environ["INDEXES_ELASTIC_LINKS"]
    index_elastic_links = index_elastic_links.split()[config.POS_ELASTIC_INDEX_LINKS]

    logging.debug("index_elastic_links", index_elastic_links)
    res_status = task_manager.create_new_index(index_elastic_links)
    logging.debug("res_status", res_status)
    if res_status != 400:
        with open(os.path.join("links_for_crawling.json"), "r", encoding="utf-8") as file:
            links_json = json.load(file)
            logging.debug(links_json)
            task_manager.add_new_links(links_json["links"])


@app.route('/health_check', methods=['GET'])
def health_check():
    check_elastic_search_status()
    return make_response(jsonify({'is_alive': 'True'}), 200)


@app.route('/task_manager/api/v1.0/add_links', methods=['POST'])
def add_links():
    print("add_links() len(request.json) -- ", len(request.json))
    if not request.json or len(request.json) == 0:
        abort(400)

    index_elastic_links = os.environ["INDEXES_ELASTIC_LINKS"]
    index_elastic_links = index_elastic_links.split()[config.POS_ELASTIC_INDEX_LINKS]
    print("add_links() index_elastic_links -- ", index_elastic_links)
    task_manager.add_new_links(request.json["links"])
    return "Links were successfully added in Elasticsearch", 201


@app.route('/task_manager/api/v1.0/get_links', methods=['GET'])
def get_links():
    check_elastic_search_status()
    return jsonify(task_manager.retrieve_links(os.environ["NUM_LINKS_PER_CRAWLER"])), 200


@app.route('/task_manager/api/v1.0/set_parsed_link_id', methods=['POST'])
def set_parsed_link_id():
    print('request.json["parsed_link_id"] -- ', request.json["parsed_link_id"])
    task_manager.set_parsed_link(int(request.json["parsed_link_id"]), 1, True)

    return "parsed_link_id was successfully set to true in Elasticsearch", 201


@app.route('/task_manager/api/v1.0/set_last_site_id', methods=['POST'])
def set_last_site_id():
    if task_manager.last_id_in_index_sites == -1:
        task_manager.last_id_in_index_sites = task_manager.get_last_site_id_in_index()

    return "last_site_id was set up", 200


@app.route('/task_manager/api/v1.0/get_last_site_id', methods=['GET'])
def get_last_site_id():
    last_id = task_manager.last_id_in_index_sites

    task_manager.last_id_in_index_sites += PACKAGE_SIZE

    print("last_id -- ", last_id)
    return jsonify({"last_site_id": last_id}), 200


if __name__ == '__main__':
    app.run(debug=True)
