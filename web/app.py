"""
Main module
"""
from langdetect import detect
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel

from elastic_interaction import elastic_search

app = FastAPI()
app.mount("/static", StaticFiles(directory="static"), name="static")
templates = Jinja2Templates(directory="templates")


@app.get('/', response_class=HTMLResponse)
async def index(request: Request):
    """
    Home page
    """
    return templates.TemplateResponse("index.html", {"request": request})


@app.get('/search', response_class=HTMLResponse)
async def search(request: Request, query: str = ""):
    """
    Search page
    """
    if not query or len(query) < 3:
        return templates.TemplateResponse("index.html", {"request": request})

    try:
        lang = detect(query)
    except Exception as e:
        print("search(): error in lang detection", e)
        lang = "uk"

    websites = await elastic_search(query, lang)
    data = {
        "query": query,
        "websites": websites
    }
    return templates.TemplateResponse("search.html", {"request": request, "data": data})


@app.post('/more_links')
async def more_links(start: int = 0, end: int = 0, search: str = ""):
    """
    Return more links
    """

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
    return {'status': 'ok', 'websites': websites}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8090)
