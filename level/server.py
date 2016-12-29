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

from importer import Importer
import tornado.ioloop
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
    importer.load()  # load all modules here

    return importer


def validate_config(config, server_parameters):
    pass


def get_context(server_parameters, config, importer):
    return Context(
        server=server_parameters,
        config=config,
        importer=importer
    )


def get_application(context):
    return context.importer.import_class(context.app_class)(context)


def run_server(application, context):
    server = HTTPServer(application)

    if context.server.fd is not None:
        fd_number = get_as_integer(context.server.fd)
        if fd_number is None:
            with open(context.server.fd, 'r') as sock:
                fd_number = sock.fileno()

        sock = socket.fromfd(fd_number,
                             socket.AF_INET | socket.AF_INET6,
                             socket.SOCK_STREAM)
        server.add_socket(sock)
    else:
        server.bind(context.server.port, context.server.host)

    server.start(1)


def run(server_parameters):
    config = get_config(server_parameters.config_path)
    validate_config(config, server_parameters)

    configure_log(config, server_parameters.log_level.upper())
    importer = get_importer(config)

    with get_context(server_parameters, config, importer) as context:
        application = get_application(context)
        run_server(application, context)

        logging.info('level running at %s:%d' % (context.server.host, context.server.port))
        tornado.ioloop.IOLoop.instance().start()
