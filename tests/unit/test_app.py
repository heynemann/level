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

from tests.unit.base import TestCase


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
