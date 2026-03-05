import http

from flask import jsonify, request

from callback.services.bytedance_asr import ByteDanceASR
from callback.utils.url_util import process_audio
from utils.log import logger
from utils.response import BizError

from . import callback_bp


@callback_bp.route("/bytedance-asr/flash", methods=["POST"])
def bytedance_asr_recognize():
    """
    【工具】字节跳动语音识别 (Flash ASR)
    ---
    tags:
      - ByteDance ASR
    summary: 语音识别
    description: 使用字节跳动火山引擎进行语音识别，resource_id 固定为 volc.bigasr.auc_turbo。
    parameters:
      - name: X-Api-App-Key
        in: header
        description: "API App Key"
        required: true
        schema:
          type: string
      - name: X-Api-Access-Key
        in: header
        description: "API Access Key"
        required: true
        schema:
          type: string
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - audio
            properties:
              audio:
                type: string
                description: 音频文件URL或base64
                example: "https://dashscope.oss-cn-beijing.aliyuncs.com/audios/welcome.mp3"
    responses:
      200:
        description: 识别成功
      400:
        description: 参数错误
      401:
        description: API Key 无效或缺失
      500:
        description: 服务端错误
    """
    data = request.get_json() or {}
    audio = data.get("audio")

    if not audio:
        raise BizError("Missing audio", code=http.HTTPStatus.BAD_REQUEST)

    audio_processed = process_audio(audio)

    app_key = request.headers.get("X-Api-App-Key")
    access_key = request.headers.get("X-Api-Access-Key")

    if not app_key or not access_key:
        raise BizError(
            "X-Api-App-Key and X-Api-Access-Key are required",
            code=http.HTTPStatus.UNAUTHORIZED,
        )

    asr = ByteDanceASR()
    result = asr.recognize(
        audio=audio_processed, app_key=app_key, access_key=access_key
    )

    return jsonify(result)
