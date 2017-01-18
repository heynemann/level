#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from importer import Importer
from preggy import expect
from tornado.testing import gen_test

from level.app import LevelApp
from level.config import Config
from level.context import Context, ServerParameters

from tests.unit.base import TestCase, WebTestCase


class AppTestCase(TestCase):
    def setUp(self):
        super(AppTestCase, self).setUp()
        self.server_parameters = ServerParameters(
            ioloop=self.io_loop,
            host='localhost',
            port=8888,
            config_path='./tests/fixtures/test-valid.conf',
            log_level='INFO',
            debug=True,
        )

        self.config = Config()
        self.importer = Importer()
        self.importer.load(
            dict(key='service_classes', module_names=self.config.SERVICES, class_name='Service'),
        )  # load all modules here

        services = []
        for service_class in self.importer.service_classes:
            srv = service_class()
            srv.name = service_class.__module__
            services.append(srv)
        self.importer.services = services
        self.context = Context(self.server_parameters, self.config, self.importer)

    @gen_test
    async def test_can_create_app(self):
        app = await LevelApp.create(self.context)
        expect(app).not_to_be_null()
        expect(app.context).to_equal(self.context)

    @gen_test
    async def test_can_initialize_services(self):
        class TestService:
            def __init__(self):
                self.initialized = False
                self.name = "TestService"
                self.app = None

            async def initialize_service(self, app):
                self.app = app
                self.initialized = True

        s = TestService()
        self.context.importer.services = [s]
        app = LevelApp(self.context, [])
        expect(app).not_to_be_null()

        await app.initialize()
        expect(s.initialized).to_be_true()
        expect(s.app).to_equal(app)

    @gen_test
    async def test_can_get_handlers_from_services(self):
        class TestService:
            def __init__(self):
                self.initialized = False
                self.name = "TestService"
                self.app = None

            async def initialize_service(self, app):
                self.app = app
                self.initialized = True

            async def get_handlers(self):
                return (
                    ('/test', None),
                )

        s = TestService()
        self.context.importer.services = [s]
        app = LevelApp(self.context, [])
        expect(app).not_to_be_null()

        handlers = await app.get_handlers()
        expect(handlers).to_length(2)
        expect(handlers[1]).to_be_like(
            ('/test', None),
        )


class WebSocketTestCase(WebTestCase):
    def setUp(self):
        super(WebSocketTestCase, self).setUp()

        class TestService:
            def __init__(self):
                self.message = None
                self.name = 'TestService'

            async def on_message(self, message):
                self.message = message
                self.user_id = message['payload']['user_id']

        self.service = TestService()

        self.context.importer.services = [self.service]

    @gen_test
    async def test_can_receive_open_message(self):
        ws = await self.websocket_connect('/ws')
        expect(ws).not_to_be_null()

        await self.wait_for(lambda: self.service.message is not None)

        expect(self.service.user_id).not_to_be_null()
        expect(self.service.message).to_be_like({
            'type': 'core.connection.open',
            'payload': {
                'user_id': self.service.user_id,
            },
        })

    @gen_test
    async def test_can_receive_close_message(self):
        ws = await self.websocket_connect('/ws')
        expect(ws).not_to_be_null()

        # wait for open
        await self.wait_for(lambda: self.service.message is not None)
        self.service.message = None

        ws.close()
        await self.wait_for(lambda: self.service.message is not None)

        expect(self.service.user_id).not_to_be_null()
        expect(self.service.message).to_be_like({
            'type': 'core.connection.close',
            'payload': {
                'user_id': self.service.user_id,
            },
        })
