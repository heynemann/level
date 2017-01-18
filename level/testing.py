#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from tornado.testing import AsyncHTTPTestCase
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
            ioloop=self.io_loop,
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
        server_parameters = self.get_server_parameters()
        conf = self.get_config()
        importer = self.get_importer(conf)
        context = self.get_context(server_parameters, conf, importer)

        app = self.io_loop.run_sync(lambda: LevelApp.create(context))

        return app

    async def fetch(self, path, **kwargs):
        response = await self.http_client.fetch(self.get_url(path), **kwargs)
        return response
