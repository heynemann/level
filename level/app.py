#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


from uuid import uuid4

import logging
from tornado.web import Application, asynchronous
from tornado.websocket import WebSocketHandler
from tornado.ioloop import IOLoop


class WSHandler(WebSocketHandler):
    def finish(self):
        pass

    @asynchronous
    async def open(self):
        self.user_id = uuid4()
        await self.application.handle_websocket_open(self.user_id)

    @asynchronous
    async def on_message(self, message):
        await self.application.handle_websocket_message(self.user_id, message)

    @asynchronous
    async def on_close(self):
        await self.application.handle_websocket_close(self.user_id)


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
        self.ioloop = IOLoop.instance()
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

    async def handle_websocket_open(self, user_id):
        await self.handle_websocket_operation('core.connection.open', {
            'type': 'core.connection.open',
            'payload': {
                'user_id': user_id,
            }
        })

    async def handle_websocket_close(self, user_id):
        await self.handle_websocket_operation('core.connection.close', {
            'type': 'core.connection.close',
            'payload': {
                'user_id': user_id,
            }
        })

    async def handle_websocket_message(self, user_id, message):
        msg_type = message['type']
        del message['type']
        await self.handle_websocket_operation(msg_type, {
            'type': msg_type,
            'payload': message,
        })

    async def handle_websocket_operation(self, method_name, *args, **kw):
        logging.debug(f'Handling {method_name} started...')
        for service in self.context.importer.services:
            method = getattr(service, 'on_message', None)
            if method is None:
                logging.debug(f"Service {service.name} does not handle websocket messages. Skipping...")
                continue

            logging.debug(f"Handling {method_name} in service {service.name}...")
            try:
                await method(*args, **kw)
                logging.debug(f"Service {service.name} handled {method_name} successfully.")
            except Exception as err:
                logging.error(f"Service {service.name} failed to handle {method_name} ({ err }).")
