#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

import mock
import asyncio

from preggy import expect

from level.app import LevelApp
from level.config import Config
import level.server
from level.server import (
    get_as_integer,
    get_config,
    configure_log,
    get_importer,
    get_context,
    get_application,
    run_server,
)

from tests.base import TestCase, async_case


class ServerTestCase(TestCase):
    def setUp(self):
        self.loop = asyncio.new_event_loop()
        asyncio.set_event_loop(None)

    def test_can_get_value_as_integer(self):
        expect(get_as_integer("1")).to_equal(1)
        expect(get_as_integer("a")).to_be_null()
        expect(get_as_integer("")).to_be_null()
        expect(get_as_integer(None)).to_be_null()

    def test_can_get_config_from_path(self):
        config = get_config('./tests/fixtures/test-valid.conf')

        expect(config).not_to_be_null()
        expect(config.RANDOM_CONFIG).to_equal('qwe')

    @mock.patch('logging.basicConfig')
    def test_can_configure_log_from_config(self, basic_config_mock):
        conf = Config()
        configure_log(conf, 'DEBUG')

        params = dict(
            datefmt='%Y-%m-%d %H:%M:%S',
            level=10,
            format='%(asctime)s %(name)s:%(levelname)s %(message)s'
        )

        basic_config_mock.assert_called_with(**params)

    @mock.patch('logging.config.dictConfig')
    def test_can_configure_log_from_dict_config(self, dict_config_mock):
        conf = Config(
            LEVEL_LOG_CONFIG={
                "level": "INFO"
            }
        )
        configure_log(conf, 'DEBUG')

        params = dict(
            level="INFO",
        )

        dict_config_mock.assert_called_with(params)

    def test_can_import_default_modules(self):
        conf = Config()
        importer = get_importer(conf)

        expect(importer).not_to_be_null()
        expect(importer.services).not_to_be_empty()

    def test_get_context(self):
        server_parameters = mock.Mock(app_class='level.app.LevelApp')
        conf = Config()
        importer = get_importer(conf)
        context = get_context(server_parameters, conf, importer)

        expect(context).not_to_be_null()

    @async_case
    async def test_get_application(self):
        server_parameters = mock.Mock(ioloop=self.loop, app_class='level.app.LevelApp')
        conf = Config()
        importer = get_importer(conf)
        context = get_context(server_parameters, conf, importer)
        app = await get_application(context)

        expect(app).not_to_be_null()
        expect(app).to_be_instance_of(LevelApp)

    @mock.patch.object(level.server, 'HTTPServer')
    @async_case
    async def test_can_run_server_with_default_params(self, server_mock):
        server_parameters = mock.Mock(host='0.0.0.0', port=1234, ioloop=self.loop, fd=None, app_class='level.app.LevelApp')
        conf = Config()
        importer = get_importer(conf)
        context = get_context(server_parameters, conf, importer)

        server_instance_mock = mock.Mock()
        server_mock.return_value = server_instance_mock

        await run_server(context)

        server_instance_mock.bind.assert_called_with(1234, '0.0.0.0')
        server_instance_mock.start.assert_called_with(1)
