from pytrends.request import TrendReq
import requests


if __name__ == '__main__':
    # pytrend = TrendReq(hl='RU')
    #
    # df = pytrend.top_charts(2020, tz=300, geo='GLOBAL')
    # print(len(df))
    # print(df)

    url = "https://api.keywordtool.io/v2/search/suggestions/google"

    data = {
        "apikey": "[APIKEY]",
        "keyword": "apple",
        "category": "web",
        "country": "US",
        "language": "en",
        "type": "suggestions",
        "exclude": [],
        "metrics": "true",
        "metrics_location": [
            2840
        ],
        "metrics_language": [
            "en"
        ],
        "metrics_network": "googlesearchnetwork",
        "metrics_currency": "USD",
        "output": "json"
    }

    response = requests.post(url, json = data)

    print(response.text.encode('utf8'))
