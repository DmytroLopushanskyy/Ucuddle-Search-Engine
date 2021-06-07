import json
import os

from flask import make_response, request, abort
from flask import Flask, jsonify, g

from task_manager import get_task_manager_singleton
from logger import get_logger
import config

app = Flask(__name__)
app.secret_key = config.FLASK_KEY
app.config['SECRET_KEY'] = config.SECRET_KEY


@app.before_first_request
def web_app_setup():
    g.task_manager = get_task_manager_singleton()
    g.logger = get_logger()
    g.logger.log("Check ES Status")
    index_elastic_links = os.environ["INDEX_ELASTIC_LINKS"]
    g.logger.log("index_elastic_links", index_elastic_links)
    res_status = g.task_manager.create_new_index(index_elastic_links)
    g.logger.log("res_status", res_status)


@app.before_request
def before_request():
    g.task_manager = get_task_manager_singleton()
    g.logger = get_logger()


@app.errorhandler(404)
def not_found(error):
    return make_response(jsonify({'error': 'Not found'}), 404)


@app.route('/health_check', methods=['GET'])
def health_check():
    return make_response(jsonify({'is_alive': 'True'}), 200)


@app.route('/task_manager/api/v1.0/add_links', methods=['POST'])
def add_links():
    if not request.json or len(request.json):
        abort(400)

    g.task_manager.add_new_links(request.json, os.environ["INDEX_ELASTIC_LINKS"])
    return "Links were successfully added in Elasticsearch", 201


@app.route('/task_manager/api/v1.0/get_links', methods=['GET'])
def get_links():
    g.logger.log("get links request")
    return jsonify(g.task_manager.retrieve_links(os.environ["NUM_LINKS_PER_CRAWLER"])), 200


if __name__ == '__main__':
    app.run(debug=True)
