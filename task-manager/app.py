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


@app.route('/health_check', methods=['GET'])
def health_check():
    check_elastic_search_status()
    return make_response(jsonify({'is_alive': 'True'}), 200)


@app.route('/task_manager/api/v1.0/add_links', methods=['POST'])
def add_links():
    if not request.json or len(request.json) == 0:
        abort(400)

    index_elastic_links = os.environ["INDEXES_ELASTIC_LINKS"]
    index_elastic_links = index_elastic_links.split()[config.POS_ELASTIC_INDEX_LINKS]
    print("add_links() index_elastic_links -- ", index_elastic_links)
    task_manager.add_new_links(request.json["links_index_name"], request.json["links"])
    return "Links were successfully added in Elasticsearch", 201


@app.route('/task_manager/api/v1.0/get_links', methods=['GET'])
def get_links():
    check_elastic_search_status()

    links = task_manager.retrieve_links(os.environ["NUM_LINKS_PER_CRAWLER"], False)
    if len(links["links"]) == 0:
        links["links"] = [["links ended", "-1"]]

    return jsonify(links), 200


@app.route('/task_manager/api/v1.0/get_failed_links', methods=['GET'])
def get_failed_links():
    check_elastic_search_status()
    links = task_manager.retrieve_links(os.environ["NUM_LINKS_PER_CRAWLER"], True)
    if len(links["links"]) == 0:
        links["links"] = [["links ended", "-1"]]

        indexes_elastic_links = os.environ["INDEXES_ELASTIC_LINKS"].split()
        if config.POS_ELASTIC_INDEX_LINKS < len(indexes_elastic_links) - 1:
            config.POS_ELASTIC_INDEX_LINKS += 1
            task_manager.index_elastic_links = indexes_elastic_links[config.POS_ELASTIC_INDEX_LINKS]

    return jsonify(links), 200


@app.route('/task_manager/api/v1.0/set_parsed_link_id', methods=['POST'])
def set_parsed_link_id():
    print('request.json["parsed_link_id"] -- ', request.json["parsed_link_id"])
    response = task_manager.set_parsed_link(int(request.json["parsed_link_id"]), 1, True)

    if response == 0:
        return "parsed_link_id was successfully set to true in Elasticsearch", 201
    else:
        return "error with setting parsed_link_id", 500


@app.route('/task_manager/api/v1.0/set_last_site_id', methods=['POST'])
def set_last_site_id():
    if task_manager.last_id_in_index_sites == -1:
        task_manager.last_id_in_index_sites = task_manager.get_last_site_id_in_all_indexes()

    return "last_site_id was set up", 200


@app.route('/task_manager/api/v1.0/get_last_site_id', methods=['GET'])
def get_last_site_id():
    last_id = task_manager.last_id_in_index_sites

    if last_id == -1:
        task_manager.last_id_in_index_sites = task_manager.get_last_site_id_in_all_indexes()
        last_id = task_manager.last_id_in_index_sites

    task_manager.last_id_in_index_sites += PACKAGE_SIZE
    print("last_site_id -- ", last_id)

    if last_id == -1:
        return jsonify({"last_site_id": False}), 500

    return jsonify({"last_site_id": last_id}), 200


if __name__ == '__main__':
    app.run(debug=True)
