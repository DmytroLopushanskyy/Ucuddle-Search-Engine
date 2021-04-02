import json
import os
import logging

from flask import make_response, request, abort
from flask import Flask, jsonify

from task_manager import TaskManager
import config

app = Flask(__name__)
app.secret_key = config.FLASK_KEY
app.config['SECRET_KEY'] = config.SECRET_KEY
task_manager = TaskManager()


# TODO: test it
@app.errorhandler(404)
def not_found(error):
    return make_response(jsonify({'error': 'Not found'}), 404)


def check_elastic_search_status():
    logging.debug("Check ES Status")
    index_elastic_links = os.environ["INDEX_ELASTIC_LINKS"]
    logging.debug("index_elastic_links", index_elastic_links)

    res_status = task_manager.create_new_index(index_elastic_links)
    logging.debug("res_status", res_status)
    if res_status != 400:
        with open(os.path.join("links_for_crawling.json"), "r", encoding="utf-8") as file:
            links_json = json.load(file)
            logging.debug(links_json)
            task_manager.add_new_links(links_json["links"], index_elastic_links)


@app.route('/health_check', methods=['GET'])
def health_check():
    check_elastic_search_status()
    return make_response(jsonify({'is_alive': 'True'}), 200)


@app.route('/task_manager/api/v1.0/add_links', methods=['POST'])
def add_links():
    if not request.json or len(request.json):
        abort(400)

    task_manager.add_new_links(request.json, os.environ["INDEX_ELASTIC_LINKS"])
    return "Links were successfully added in Elasticsearch", 201


@app.route('/task_manager/api/v1.0/get_links', methods=['GET'])
def get_links():
    check_elastic_search_status()
    return jsonify(task_manager.retrieve_links(os.environ["NUM_LINKS_PER_CRAWLER"])), 200


if __name__ == '__main__':
    check_elastic_search_status()
    app.run(debug=True)
