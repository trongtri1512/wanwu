import os
import sys
from typing import Optional

import requests

current_dir = os.path.dirname(os.path.abspath(__file__))
parent_dir = os.path.dirname(os.path.dirname(current_dir))

if parent_dir not in sys.path:
    sys.path.append(parent_dir)

from utils.log import logger


class AliASR:
    API_URL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
    MODEL = "qwen3-asr-flash"

    def recognize(self, audio: str, api_key: str, model: str = MODEL):
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
            "Connection": "keep-alive",
        }

        payload = {
            "model": model,
            "input": {"messages": [{"content": [{"audio": audio}], "role": "user"}]},
        }

        try:
            response = requests.post(
                self.API_URL, headers=headers, json=payload, timeout=30
            )
            response.raise_for_status()
            result = response.json()
            return result
        except requests.exceptions.RequestException as e:
            logger.error(f"Ali ASR request failed: {e}")
            return {"error": str(e)}
