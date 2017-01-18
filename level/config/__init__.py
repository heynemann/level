#!/usr/bin/env python
# -*- coding: utf-8 -*-

# this file is part of level.
# https://github.com/heynemann/level

# licensed under the mit license:
# http://www.opensource.org/licenses/mit-license
# copyright (c) 2016, bernardo heynemann <heynemann@gmail.com>


import derpconf.config as config
from derpconf.config import Config

Config.define('LEVEL_LOG_CONFIG', None, 'Logging configuration as json', 'Logging')
Config.define(
    'LEVEL_LOG_FORMAT', '%(asctime)s %(name)s:%(levelname)s %(message)s',
    'Log Format to be used by level when writing log messages.', 'Logging'
)

Config.define(
    'LEVEL_LOG_DATE_FORMAT', '%Y-%m-%d %H:%M:%S',
    'Date Format to be used by level when writing log messages.', 'Logging'
)

Config.define(
    'APP_CLASS', 'level.app.LevelApp',
    'Custom app class to override LevelApp.', 'WebServer',
)

Config.define(
    'WS_URL', '/ws',
    'URL for Websocket Access', 'WebServer',
)


Config.define(
    'SERVICES', (
        'level.services.healthcheck',
    ), 'List of services to be used in this level server', 'Services'
)


def generate_config():
    config.generate_config()


if __name__ == '__main__':
    generate_config()
