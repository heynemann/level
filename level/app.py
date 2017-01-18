#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


import traceback
from uuid import uuid4

import logging
from tornado.web import Application, asynchronous
from tornado.websocket import WebSocketHandler
from tornado.ioloop import IOLoop

from level.json import loads


class WSHandler(WebSocketHandler):
    def finish(self):
        pass

    @asynchronous
    async def open(self):
        self.socket_id = str(uuid4())
        self.application.connected_players[self.socket_id] = self
        await self.application.handle_websocket_open(self.socket_id)

    @asynchronous
    async def on_message(self, message):
        msg = loads(message)
        await self.application.handle_websocket_message(self.socket_id, msg)

    @asynchronous
    async def on_close(self):
        del self.application.connected_players[self.socket_id]
        await self.application.handle_websocket_close(self.socket_id)

    @asynchronous
    async def send_to_socket(self, msg):
        await self.write_message(msg)


class LevelApp(Application):
    @classmethod
    async def create(cls, context, *args, **kw):
        logging.debug('Creating new application...')
        app = cls(context, [])
        await app.initialize()
        handlers = await app.get_handlers()
        if context.server.debug:
            kw['autoreload'] = True
        app = cls(context, handlers, *args, **kw)
        logging.debug('Application created successfully.')
        return app

    def __init__(self, context, handlers, *args, **kw):
        self.context = context
        self.io_loop = IOLoop.instance()
        self.connected_players = {}
        super(LevelApp, self).__init__(handlers, *args, **kw)

    async def initialize(self):
        logging.debug('Initializing application...')
        for service in self.context.importer.services:
            logging.debug(f"Initializing service {service.name}...")
            await service.initialize_service(self)
            logging.debug(f"Service {service.name} initialized successfully.")

    async def get_handlers(self):
        logging.debug('Loading HTTP Handlers...')
        handlers = [
            (self.context.config.WS_URL, WSHandler),
        ]

        for service in self.context.importer.services:
            logging.debug(f"Retrieving HTTP handlers for service {service.name}...")
            service_handlers = await service.get_handlers()
            handlers += list(service_handlers)
            logging.debug(f"HTTP handlers for service {service.name} retrieved successfully.")

        logging.debug(f"HTTP Handlers loaded successfully ({len(handlers)} handlers loaded).")
        logging.debug(handlers)
        return tuple(handlers)

    async def handle_websocket_open(self, socket_id):
        await self.handle_websocket_operation({
            'type': 'core.connection.open',
            'socket_id': socket_id,
            'payload': {},
        })

    async def handle_websocket_close(self, socket_id):
        await self.handle_websocket_operation({
            'type': 'core.connection.close',
            'socket_id': socket_id,
            'payload': {},
        })

    async def handle_websocket_message(self, socket_id, message):
        msg_type = message['type']
        del message['type']
        await self.handle_websocket_operation({
            'type': msg_type,
            'socket_id': socket_id,
            'payload': message,
        })

    async def handle_websocket_operation(self, msg):
        method_name = msg['type']
        logging.debug(f'Handling {method_name} started...')
        for service in self.context.importer.services:
            logging.debug(f"Handling {method_name} in service {service.name}...")
            try:
                await service.on_message(msg)
                logging.debug(f"Service {service.name} handled {method_name} successfully.")
            except Exception as err:
                logging.error(f"Service {service.name} failed to handle {method_name} ({ traceback.format_exc() }).")
