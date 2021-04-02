"""
Main module
"""
from flask import Flask, render_template, request, jsonify

import config
from elastic_interaction import elastic_search

app = Flask(__name__)
app.secret_key = config.FLASK_KEY
app.config['SECRET_KEY'] = config.SECRET_KEY


@app.route('/', methods=['GET'])
def index():
    """
    Home page
    """
    return render_template('index.html')


@app.route('/search', methods=['GET'])
def search():
    """
    Search page
    """
    query = request.args.get('query')

    websites = elastic_search(query)

    data = {
        "query": query,
        "websites": websites
    }

    if len(websites) == 0:
        return render_template('search_not_found.html', data=data)

    return render_template('search.html', data=data)


@app.route('/more_links', methods=['POST'])
def more_links():
    """
    Return more links
    """
    start = request.json['start']
    end = request.json['end']
    search = request.json['search']

    websites = [
        {
            "title": "Погода у Львові. Прогноз погоди Львів на ... - SINOPTIK",
            "link": "https://ua.sinoptik.ua/%D0%BF%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0-%D0%BB%D1%8C%D0%B2%D1%96%D0%B2",
            "description": "Погода у Львові на тиждень. Прогноз погоди у Львові. Детальний метеопрогноз у Львові, Львівська область на сьогодні, завтра, вихідні."
        },
        {
            "title": "Погода у Львові сьогодні, прогноз погоди Львів ... - GISMETEO",
            "link": "https://www.gismeteo.ua/ua/weather-lviv-4949/",
            "description": "Погода у Львові на сьогодні, точний прогноз погоди на сьогодні для населеного пункту Львів, Львів, Львівська область, Україна."
        },
        {
            "title": "METEOPROG: Погода в Львові на сьогодні, завтра, на 3 дні ...",
            "link": "https://www.meteoprog.ua/ua/weather/Lviv/",
            "description": "Погода в Львові на сьогодні и завтра, на 3 дні, на 5 днів, Львівська область. Точний прогноз погоди в Львові на METEOPROG.UA. Докладний ..."
        },
        {
            "title": "Погода у Львові (аеропорт) - РП5 - Rp5.ua",
            "link": "https://rp5.ua/%D0%9F%D0%BE%D0%B3%D0%BE%D0%B4%D0%B0_%D1%83_%D0%9B%D1%8C%D0%B2%D0%BE%D0%B2%D1%96_(%D0%B0%D0%B5%D1%80%D0%BE%D0%BF%D0%BE%D1%80%D1%82)",
            "description": "У Львові (аеропорт) завтра очікується 0..-5 °C, переважно без опадів, слабкий вітер. Післязавтра: -2..+1 °C, без опадів, свіжий вітер. РП5."
        },
        {
            "title": "Погода в Львові, Прогноз погоди Львів, Львівський район ...",
            "link": "https://pogoda.meta.ua/ua/Lvivska/Lvivskiy/Lviv/",
            "description": "Довгостроковий прогноз погоди у місті Львові на сьогодні ☔️ і найближчий місяць - стежте за температурою, опадами і вітром в Львові, Львівський ..."
        },
        {
            "title": "Український Гідрометцентр. Прогноз погоди Львів",
            "link": "https://meteo.gov.ua/ua/33393",
            "description": "Якісний прогноз погоди по м. Львів від синоптиків Українського Гідрометцентру. Поточна погода - метеодані метеорологічних, аерологічних станцій, ..."
        }
    ]
    return jsonify({'status': 'ok', 'websites': websites})


if __name__ == '__main__':
    app.run(debug=True, port=8000)
