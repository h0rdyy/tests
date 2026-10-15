from __future__ import annotations

import json
import re
import smtplib 
from email.mime.text import MIMEText 
import ssl
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional
from jinja2 import TemplateNotFound
from flask import Flask, abort, jsonify, render_template, request
import os
from flask_limiter import Limiter
from flask_limiter.util import get_remote_address
import logging
from logging.handlers import RotatingFileHandler
from dotenv import load_dotenv
load_dotenv()

app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = 1 * 1024 * 1024  # 1 MB максимум

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)

formatter = logging.Formatter("%(asctime)s [%(levelname)s] %(message)s")

file_handler = RotatingFileHandler(
    "logs/app.log",
    maxBytes=5 * 1024 * 1024,
    backupCount=3,
    encoding="utf-8"
)
file_handler.setFormatter(formatter)

console_handler = logging.StreamHandler()
console_handler.setFormatter(formatter)

root_logger = logging.getLogger()
root_logger.setLevel(logging.INFO)
root_logger.addHandler(file_handler)
root_logger.addHandler(console_handler)

logging.getLogger("werkzeug").addHandler(file_handler)
logger = logging.getLogger(__name__)

limiter = Limiter(
    get_remote_address,
    app=app,
    default_limits=["100 per hour"],
    storage_uri="memory://"  
)

SMTP_SERVER = "smtp.mail.ru"
SMTP_PORT = 465
SENDER_EMAIL = os.getenv("SENDER_EMAIL")
SENDER_PASSWORD = os.getenv("SENDER_PASSWORD")
RECEIVER_EMAIL = os.getenv("RECEIVER_EMAIL")

def send_email_notification(phone: str, source: str) -> None:
    subject = "Новая заявка с сайта!"
    body = f"У вас новая заявка!\n\nТелефон: {phone}\nФорма: {source}\nВремя: {datetime.now().strftime('%d.%m.%Y %H:%M')}"
    
    msg = MIMEText(body, "plain", "utf-8")
    msg["Subject"] = subject
    msg["From"] = SENDER_EMAIL
    msg["To"] = RECEIVER_EMAIL

    context = ssl.create_default_context()

    try:
        with smtplib.SMTP_SSL(SMTP_SERVER, SMTP_PORT, context=context) as server:
            server.login(SENDER_EMAIL, SENDER_PASSWORD)
            server.sendmail(SENDER_EMAIL, RECEIVER_EMAIL, msg.as_string())
    except Exception as e:
        logger.error("Ошибка отправки email: %s", e)

SERVICES = [
    {
        "id": 1,
        "title": "Защита персональных данных",
        "hint": "Ваш юридический щит и репутационный актив",
        "cell": "col1",
        "thumb": "shield.png",
    },
    {
        "id": 2,
        "title": "Услуги по обеспечению информационной безопасности",
        "hint": "Технологическая крепость вашего бизнеса",
        "cell": "top",
        "thumb": "fire.png",
    },
    {
        "id": 3,
        "title": "Информационное право",
        "hint": "",
        "cell": "bl",
        "thumb": "lightning.png",
    },
    {
        "id": 4,
        "title": "Разработка образовательных программ",
        "hint": "",
        "cell": "br",
        "thumb": "light.png",
    },
    {
        "id": 5,
        "title": "Обучение и развитие персонала",
        "hint": "Инвестиции в вашу безопасность и репутацию",
        "cell": "col4",
        "thumb": "study.png",
    },
]


@app.context_processor
def inject_services():
    return {"services": SERVICES}


BASE_DIR = Path(__file__).resolve().parent
DATA_DIR = BASE_DIR / "data"
LEADS_FILE = DATA_DIR / "leads.json"


def _ensure_data_dir() -> None:
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    if not LEADS_FILE.exists():
        LEADS_FILE.write_text("[]", encoding="utf-8")


def _normalize_phone(raw: str) -> Optional[str]:
    digits = re.sub(r"\D", "", raw or "")
    if len(digits) < 10:
        return None
    if len(digits) == 10:
        return "+7" + digits
    if digits.startswith("8") and len(digits) == 11:
        return "+7" + digits[1:]
    if digits.startswith("7") and len(digits) == 11:
        return "+" + digits
    if raw.strip().startswith("+") and len(digits) >= 10:
        return "+" + digits
    return None


@app.route("/")
def index():
    return render_template("index.html")


@app.route("/about")
def about():
    return render_template("about.html")

@app.route('/privacy')
def privacy_page():
    return render_template('privacy.html')

@app.route("/services/<int:service_id>")
def service_page(service_id: int):
    svc = next((s for s in SERVICES if s["id"] == service_id), None)
    if svc is None:
        abort(404)
    return render_template("service.html", service=svc)

@app.route('/expertise')
def expertise_page():
    return render_template('expertise.html', services=SERVICES) 

ALLOWED_ARTICLES = {
    "kibernsk": "articles/kibernsk.html",
    # Сюда добавляй новые статьи по мере публикации
}

@app.route('/articles/<slug>')
def article_page(slug):
    template = ALLOWED_ARTICLES.get(slug)
    if not template:
        abort(404)
    return render_template(template)

@app.errorhandler(404)
def not_found(_e):
    return render_template("404.html"), 404

@limiter.limit("5 per minute")
@app.post("/api/lead")
def api_lead():
    data = request.get_json(silent=True) or {}
    phone_raw = (data.get("phone") or "").strip()
    phone = _normalize_phone(phone_raw)
    if not phone:
        return jsonify({"ok": False, "error": "Укажите корректный номер телефона"}), 400

    _ensure_data_dir()
    try:
        existing = json.loads(LEADS_FILE.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        existing = []

    source = (data.get("source") or "unknown")[:120]

    entry = {
        "phone": phone,
        "created_at": datetime.now(timezone.utc).isoformat(),
        "source": source,
    }
    existing.append(entry)
    LEADS_FILE.write_text(json.dumps(existing, ensure_ascii=False, indent=2), encoding="utf-8")

    send_email_notification(phone, source)

    return jsonify({"ok": True})

if __name__ == "__main__":
    app.run(debug=False, port=5000)


