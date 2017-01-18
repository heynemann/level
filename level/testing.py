#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from tornado import gen
from tornado.testing import AsyncHTTPTestCase
from tornado.httpclient import HTTPRequest
from tornado.websocket import websocket_connect
from importer import Importer

from level.app import LevelApp
from level.config import Config
from level.context import ServerParameters, Context


class LevelTestCase(AsyncHTTPTestCase):
    def get_config(self):
        return Config()

    def get_importer(self, config):
        importer = Importer()
        importer.load(
            dict(key='service_classes', module_names=config.SERVICES, class_name='Service'),
        )  # load all modules here

        services = []
        for service_class in importer.service_classes:
            srv = service_class()
            srv.name = service_class.__module__
            services.append(srv)
        importer.services = services

        return importer

    def get_server_parameters(self):
        return ServerParameters(
            io_loop=self.io_loop,
            host='localhost',
            port=self.get_http_port(),
            config_path='./tests/fixtures/test-valid.conf',
            log_level='INFO',
        )

    def get_context(self, server_parameters, config, importer):
        return Context(
            server_parameters, config, importer,
        )

    def get_app(self):
        self.server_parameters = self.get_server_parameters()
        self.config = self.get_config()
        self.importer = self.get_importer(self.config)
        self.context = self.get_context(self.server_parameters, self.config, self.importer)

        self.app = self.io_loop.run_sync(lambda: LevelApp.create(self.context))

        return self.app

    async def sleep(self, time):
        await gen.sleep(time)

    async def wait_moment(self):
        await gen.moment

    async def wait_for(self, f):
        while not f():
            await gen.moment

    async def fetch(self, path, **kwargs):
        response = await self.http_client.fetch(self.get_url(path), **kwargs)
        return response

    # http://stackoverflow.com/questions/33264427/how-to-test-that-tornado-read-message-got-nothing-to-read
    async def websocket_connect(self, path):
        request = HTTPRequest(self.get_url(path).replace('http://', 'ws://'))
        self.ws = await websocket_connect(request)
        await self.wait_moment()

    def websocket_close(self):
        if self.ws is None:
            return

        self.ws.close()

        self.ws = None

    async def write_ws_message(self, message):
        await self.ws.write_message(message)
        await self.wait_moment()

    async def read_ws_message(self):
        response = await self.ws.read_message()
        await self.wait_moment()
        return response
