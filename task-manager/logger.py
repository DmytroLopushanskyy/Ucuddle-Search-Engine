import requests
from datetime import datetime
import logging
import os


class _TelegramLogger:
    _instance = None
    bot_token = os.environ['LOGGING_BOT_TOKEN']
    users = ["138918380"]

    def __init__(self):
        self.logger = logging.getLogger('dev')

    def log(self, *message):
        message = [str(item) for item in message]
        message = ", ".join(message)
        logging.debug(message)
        dt_string = datetime.now().strftime("%d/%m/%Y %H:%M:%S")
        message = dt_string + ": " + message
        for user in self.users:
            requests.get("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s" %
                (self.bot_token, user, message))


def get_logger():
    if _TelegramLogger._instance is None:
        _TelegramLogger._instance = _TelegramLogger()
    return _TelegramLogger._instance