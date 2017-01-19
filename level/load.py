#!/usr/bin/env python
# -*- coding: utf-8 -*-

# this file is part of level.
# https://github.com/heynemann/level

# licensed under the mit license:
# http://www.opensource.org/licenses/mit-license
# copyright (c) 2016, bernardo heynemann <heynemann@gmail.com>

import time

from tornado.ioloop import IOLoop
from tornado.websocket import websocket_connect
from tornado import gen

class Result:
    def __init__(self):
        self.websockets_opened = 0
        self.websockets_closed = 0
        self.start = time.time()


class LoadTest:
    def __init__(self, ws_url, concurrency, recycle, duration):
        self.ws_url = ws_url
        self.concurrency = concurrency
        self.recycle = recycle
        self.duration = duration
        self.loop = IOLoop()
        self.result = None

    def print_summary(self):
        print(f"Total of websockets opened: {self.result.websockets_opened}")
        print(f"Total of websockets closed: {self.result.websockets_closed}")

    async def websocket_connect(self):
        ws = await websocket_connect(self.ws_url)
        await gen.moment
        self.result.websockets_opened += 1
        self.loop.call_later(self.recycle, self.recycle_ws(ws))

    def recycle_ws(self, ws):
        async def handle():
            ws.close()
            self.result.websockets_closed += 1
            await self.websocket_connect()
        return handle

    def start(self):
        self.loop.make_current()
        self.result = Result()

        for _ in range(self.concurrency):
            self.loop.call_later(0, self.websocket_connect)

        self.loop.call_later(self.duration, self.stop)
        self.loop.start()

    def stop(self):
        self.print_summary()
        self.loop.stop()
