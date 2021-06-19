import json
from langdetect import detect


if __name__ == '__main__':
    with open("../files/ua_domains_670k.txt", "r", encoding="utf-8") as f:
        file_lines = f.readlines()

    domains = dict()
    domains["links"] = []
    len_file_lines = len(file_lines)
    for i in range(10000, len_file_lines):
        domains["links"].append(file_lines[i].strip())

    with open("../files/ua_domains_660k.json", "w", encoding="utf-8") as outfile:
        json.dump(domains, outfile, indent=4)
