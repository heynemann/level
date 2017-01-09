#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from preggy import expect

from level.json import loads, dumps
from tests.base import TestCase


class JSONTestCase(TestCase):
    def test_can_dump(self):
        expect(dumps({
            "x": 1,
        })).to_equal('{"x":1}')

    def test_can_load(self):
        expect(loads('{"x":1}')).to_be_like({
            "x": 1,
        })
