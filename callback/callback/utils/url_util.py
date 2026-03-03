import re
from urllib.parse import urlparse

import requests

from configs.config import config
from utils.log import logger


def is_url(value: str) -> bool:
    try:
        result = urlparse(value)
        # 确保有协议 (scheme) 和 域名 (netloc)
        return all([result.scheme in ["http", "https"], result.netloc])
    except:
        return False


def get_mime_type(url: str) -> str:
    url_lower = url.lower()
    if url_lower.endswith(".mp3") or url_lower.endswith(".mpeg"):
        return "audio/mpeg"
    elif url_lower.endswith(".wav"):
        return "audio/wav"
    elif url_lower.endswith(".ogg"):
        return "audio/ogg"
    elif url_lower.endswith(".flac"):
        return "audio/flac"
    elif url_lower.endswith(".aac"):
        return "audio/aac"
    elif url_lower.endswith(".m4a"):
        return "audio/mp4"
    elif url_lower.endswith(".wma"):
        return "audio/x-ms-wma"
    elif url_lower.endswith(".aiff"):
        return "audio/aiff"
    return "audio/mpeg"


def url_to_base64(url: str) -> str:
    try:
        url_to_base64_api = config.callback_cfg["WANWU"]["CALLBACK_URL_TO_BASE64"]
    except (KeyError, TypeError):
        url_to_base64_api = ""

    if not url_to_base64_api:
        logger.warning("CALLBACK_URL_TO_BASE64 not configured, returning original URL")
        return url

    try:
        response = requests.post(url_to_base64_api, json={"fileUrl": url}, timeout=30)
        response.raise_for_status()
        result = response.json()

        if result.get("code") == 0:
            return result.get("data", url)

        logger.error(f"URL to base64 failed: {result}")
        return url
    except Exception as e:
        logger.error(f"URL to base64 request failed: {e}")
        return url


def url_to_base64_with_mime(url: str) -> str:
    try:
        url_to_base64_api = config.callback_cfg["WANWU"]["CALLBACK_URL_TO_BASE64"]
    except (KeyError, TypeError):
        url_to_base64_api = ""

    if not url_to_base64_api:
        logger.warning("CALLBACK_URL_TO_BASE64 not configured, returning original URL")
        return url

    try:
        mime_type = get_mime_type(url)

        payload = {
            "fileUrl": url,
            "addPrefix": True,
            "customPrefix": f"data:{mime_type};base64,",
        }

        response = requests.post(url_to_base64_api, json=payload, timeout=30)
        response.raise_for_status()
        result = response.json()

        if result.get("code") == 0:
            return result.get("data", url)

        logger.error(f"URL to base64 failed: {result}")
        return url
    except Exception as e:
        logger.error(f"URL to base64 request failed: {e}")
        return url


def process_audio(audio: str) -> str:
    if is_url(audio):
        return url_to_base64(audio)
    return audio


def process_audio_mime(audio: str) -> str:
    if is_url(audio):
        return url_to_base64_with_mime(audio)
    return audio
