#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from uuid import uuid4

from preggy import expect
from tornado import gen

from level.config import Config
from level.extensions.pubsub.redis import PubSub

from tests.base import TestCase, async_case


class RedisPubSubTestCase(TestCase):
    def setUp(self):
        self.config = Config(
            REDIS_HOST='localhost',
            REDIS_PORT=4444,
            REDIS_DATABASE=0,
            REDIS_PASSWORD=None,
        )

    def test_can_create_instance(self):
        ps = PubSub(self.config)
        expect(ps).not_to_be_null()

    @async_case
    async def test_can_initialize(self):
        ps = PubSub(self.config)
        await ps.initialize()
        expect(ps.redis).not_to_be_null()

    @async_case
    async def test_can_cleanup(self):
        ps = PubSub(self.config)
        await ps.initialize()
        expect(ps.redis).not_to_be_null()

        await ps.cleanup()
        expect(ps.redis_closed).to_be_true()
        expect(ps.redis).to_be_null()
        expect(ps.subs_closed).to_be_true()
        expect(ps.subs).to_be_null()

    @async_case
    async def test_can_subscribe_to_channel(self):
        ps = PubSub(self.config)
        await ps.initialize()
        expect(ps.redis).not_to_be_null()

        received = {}

        def on_message(chan, msg):
            received[chan] = msg

        chan = str(uuid4())
        await ps.subscribe(chan, on_message)

        await ps.publish(chan, 'qwe')

        while chan not in received:
            await gen.sleep(0.001)

        expect(received).to_include(chan)
        expect(received[chan]).to_be_like('qwe')
