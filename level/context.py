#!/usr/bin/env python
# -*- coding: utf-8 -*-

# this file is part of level.
# https://github.com/heynemann/level

# licensed under the mit license:
# http://www.opensource.org/licenses/mit-license
# copyright (c) 2016, bernardo heynemann <heynemann@gmail.com>


class Context:
    '''
    Class responsible for containing:
    * Server parameters;
    * Configurations read from config file (or defaults);
    * Importer with imported modules.
    '''

    def __init__(self, server=None, config=None, importer=None):
        self.server = server
        self.config = config
        self.importer = importer

        self.app_class = 'level.app.LevelApp'

        if hasattr(self.config, 'APP_CLASS'):
            self.app_class = self.config.APP_CLASS

    def __enter__(self):
        return self

    def __exit__(self, type, value, traceback):
        pass


class ServerParameters(object):
    def __init__(self, io_loop, host, port, config_path, log_level, debug=False, fd=None):
        self.io_loop = io_loop
        self.host = host
        self.port = port
        self.config_path = config_path
        self.log_level = log_level
        self.debug = debug
        self.fd = fd
