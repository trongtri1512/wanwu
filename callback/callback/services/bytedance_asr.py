import os
import sys
import uuid
from typing import Optional

import requests

current_dir = os.path.dirname(os.path.abspath(__file__))
parent_dir = os.path.dirname(os.path.dirname(current_dir))

if parent_dir not in sys.path:
    sys.path.append(parent_dir)

from utils.log import logger


class ByteDanceASR:
    RESOURCE_ID = "volc.bigasr.auc_turbo"
    API_URL = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/recognize/flash"

    def recognize(self, audio: str, app_key: str, access_key: str):
        request_id = str(uuid.uuid4())

        headers = {
            "X-Api-App-Key": app_key,
            "X-Api-Access-Key": access_key,
            "X-Api-Resource-Id": self.RESOURCE_ID,
            "X-Api-Request-Id": request_id,
            "X-Api-Sequence": "-1",
            "Content-Type": "application/json",
        }

        payload = {"audio": {"data": audio}}

        try:
            response = requests.post(
                self.API_URL, headers=headers, json=payload, timeout=30
            )
            response.raise_for_status()
            result = response.json()
            return result
        except requests.exceptions.RequestException as e:
            logger.error(f"ByteDance ASR request failed: {e}")
            return {"error": str(e)}
