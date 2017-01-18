#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from preggy import expect
from tornado.testing import gen_test

from level.json import loads
from tests.base import WebTestCase


class HealthcheckTestCase(WebTestCase):
    @gen_test
    async def test_can_get_healthcheck(self):
        response = await self.fetch('/healthcheck')
        expect(response).not_to_be_null()
        expect(response.code).to_equal(200)

        obj = loads(response.body)
        expect(obj['success']).to_be_true()
