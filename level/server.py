#!/usr/bin/env python
# -*- coding: utf-8 -*-

# this file is part of level.
# https://github.com/heynemann/level

# licensed under the mit license:
# http://www.opensource.org/licenses/mit-license
# copyright (c) 2016, bernardo heynemann <heynemann@gmail.com>


import logging
import logging.config
import os
import socket
from os.path import expanduser, dirname, join
import asyncio

from importer import Importer
from tornado.httpserver import HTTPServer

from level.config import Config
from level.context import Context


def get_as_integer(value):
    try:
        return int(value)
    except (ValueError, TypeError):
        return None


def get_config(config_path):
    lookup_paths = [os.curdir,
                    expanduser('~'),
                    '/etc/level/',
                    join(dirname(__file__), 'level', 'config')]

    return Config.load(config_path, conf_name='level.conf', lookup_paths=lookup_paths)


def configure_log(config, log_level):
    if (config.LEVEL_LOG_CONFIG and config.LEVEL_LOG_CONFIG != ''):
        logging.config.dictConfig(config.LEVEL_LOG_CONFIG)
    else:
        logging.basicConfig(
            level=getattr(logging, log_level),
            format=config.LEVEL_LOG_FORMAT,
            datefmt=config.LEVEL_LOG_DATE_FORMAT
        )


def get_importer(config):
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


def validate_config(config, server_parameters):
    pass


def get_context(server_parameters, config, importer):
    return Context(
        server=server_parameters,
        config=config,
        importer=importer
    )


async def get_application(context):
    return await context.importer.import_class(context.app_class).create(context)


async def run_server(context):
    application = await get_application(context)
    server = HTTPServer(application)

    server.bind(context.server.port, context.server.host)

    server.start(1)


def run(server_parameters):
    config = get_config(server_parameters.config_path)
    configure_log(config, server_parameters.log_level.upper())
    importer = get_importer(config)

    with get_context(server_parameters, config, importer) as context:
        validate_config(context)
        logging.info('level running at %s:%d' % (context.server.host, context.server.port))
        asyncio.ensure_future(run_server(context), loop=server_parameters.ioloop)
        asyncio.get_event_loop().run_forever()
