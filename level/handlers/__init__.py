#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


import logging
import tornado.web

from level.json import dumps


class BaseHandler(tornado.web.RequestHandler):
    # def prepare(self, *args, **kwargs):
        # super(BaseHandler, self).prepare(*args, **kwargs)

    # def on_finish(self, *args, **kwargs):
        # super(BaseHandler, self).on_finish(*args, **kwargs)

    def _error(self, status, msg=None):
        self.set_status(status)
        if msg is not None:
            logging.warn(msg)
        self.finish()

    def json(self, obj, status=200):
        self.set_status(status)
        self.write(dumps(obj))
