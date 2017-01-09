#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


import logging
from tornado.web import Application


class LevelApp(Application):
    @classmethod
    async def create(cls, context, *args, **kw):
        logging.debug('Creating new application...')
        app = cls(context, [])
        await app.initialize()
        handlers = await app.get_handlers()
        app = cls(context, handlers, *args, **kw)
        logging.debug('Application created successfully.')
        return app

    def __init__(self, context, handlers, *args, **kw):
        self.context = context
        super(LevelApp, self).__init__(handlers, *args, **kw)

    async def initialize(self):
        logging.debug('Initializing application...')
        for service in self.context.importer.services:
            logging.debug(f"Initializing service {service.name}...")
            await service.initialize_service(self)
            logging.debug(f"Service {service.name} initialized successfully.")

    async def get_handlers(self):
        logging.debug('Loading HTTP Handlers...')
        handlers = tuple()

        for service in self.context.importer.services:
            logging.debug(f"Retrieving HTTP handlers for service {service.name}...")
            handlers += await service.get_handlers()
            logging.debug(f"HTTP handlers for service {service.name} retrieved successfully.")

        logging.debug(f"HTTP Handlers loaded successfully ({len(handlers)} handlers loaded).")
        return handlers
