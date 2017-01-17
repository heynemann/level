#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from collections import defaultdict

from tornado import gen
import toredis


class PubSub:
    def __init__(self, config):
        self.config = config
        self.subs_closed = True
        self.subs = None
        self.redis_closed = True
        self.redis = None
        self.handlers = defaultdict(list)
        self.__handler_subscribed = {}

    # Wrap redis operations in tasks
    def __perform_async(self, method):
        def perform(*arg, **kw):
            return gen.Task(method, *arg, **kw)
        return perform

    # If method not found get redis command instead
    def __getattr__(self, key):
        if key in self.__dict__:
            return super(PubSub, self).__getattr__(key)
        return self.__perform_async(getattr(self.redis, key))

    def define_configuration_defaults(self):
        self.config.define('REDIS_HOST', 'localhost', 'Redis PubSub host', 'Redis Extension')
        self.config.define('REDIS_PORT', 6379, 'Redis PubSub port', 'Redis Extension')
        self.config.define('REDIS_DATABASE', 0, 'Redis PubSub database', 'Redis Extension')
        self.config.define('REDIS_PASSWORD', None, 'Redis PubSub database', 'Redis Extension')

    async def initialize(self):
        redis_host = self.config.REDIS_HOST
        redis_port = self.config.REDIS_PORT
        redis_db = self.config.REDIS_DATABASE
        redis_pass = self.config.REDIS_PASSWORD

        self.subs = ToRedisClient()
        await self.subs.connect(
            host=redis_host,
            port=redis_port,
            database=redis_db,
            password=redis_pass,
        )
        self.subs_closed = False

        self.redis = ToRedisClient()
        await self.redis.connect(
            host=redis_host,
            port=redis_port,
            database=redis_db,
            password=redis_pass,
        )
        self.redis_closed = False

    async def cleanup(self):
        if self.subs is not None:
            self.subs.close()
            self.subs = None
        self.subs_closed = True

        if self.redis is not None:
            self.redis.close()
            self.redis = None
        self.redis_closed = True

    def pub_handler(self, response):
        res_type, channel, payload = response
        channel = channel.decode('utf-8')
        if res_type == b'subscribe':
            self.__handler_subscribed[channel] = True

        if res_type == b'message':
            for handler in self.handlers[channel]:
                self.subs._io_loop.call_later(0, lambda: handler(channel, payload))

    async def subscribe(self, channel, handler):
        first_sub = channel not in self.handlers
        self.handlers[channel].append(handler)
        if first_sub:
            self.__handler_subscribed[channel] = False
            self.subs.subscribe(channel, self.pub_handler)

            while not self.__handler_subscribed[channel]:
                # print(self.__handler_subscribed)
                await gen.sleep(0.01)

    async def publish(self, channel, payload):
        return await gen.Task(self.redis.publish, channel, payload)


class ToRedisClient(toredis.Client):
    def on_disconnect(self):
        if self.auto_reconnect:
            self._io_loop.call_later(1, self.reconnect)

    async def connect(
            self,
            host='localhost',
            port=6379,
            password=None,
            database=0,
            auto_reconnect=True):
        self.host = host
        self.port = port
        self.password = password
        self.database = database
        self.auto_reconnect = auto_reconnect
        await self.reconnect()

    async def reconnect(self):
        conn_method = super(ToRedisClient, self).connect
        await gen.Task(conn_method, self.host, self.port)
        await self.auth_first()

    async def auth_first(self):
        # Select database
        status = await gen.Task(self.select, self.database)
        assert status == b'OK', status

        if self.password is None:
            return

        # Authenticate first
        self.status = await gen.Task(self.auth, self.password)
