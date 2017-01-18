#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from unittest import TestCase as PythonTestCase

from tornado.ioloop import IOLoop

from level.testing import LevelTestCase


def async_case(f, *args, **kw):
    def handle_method(*args, **kw):
        async def go():
            await f(*args, **kw)

        loop = IOLoop.instance()
        loop.run_sync(go)

    handle_method.__name__ = f.__name__
    return handle_method


class TestCase(PythonTestCase):
    def setUp(self):
        super(TestCase, self).setUp()
        self.io_loop = IOLoop()
        self.io_loop.make_current()


class WebTestCase(LevelTestCase):
    pass
