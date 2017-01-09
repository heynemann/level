#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

import asyncio
from unittest import TestCase as PythonTestCase


def async_case(f, *args, **kw):
    def handle_method(*args, **kw):
        async def go():
            await f(*args, **kw)

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(None)
        loop.run_until_complete(go())

    handle_method.__name__ = f.__name__
    return handle_method


class TestCase(PythonTestCase):
    pass
