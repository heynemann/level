#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


from tornado.web import Application

from level.handlers.healthcheck import HealthcheckHandler


class LevelApp(Application):
    def __init__(self, *args, **kw):
        super(LevelApp, self).__init__(self.get_handlers(), *args, **kw)

    def get_handlers(self):
        return [
            (r'/healthcheck', HealthcheckHandler),
        ]
