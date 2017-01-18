#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


from level.json import dumps


class BaseService:
    services = []

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        cls.services.append(cls)

    async def initialize_service(self, app):
        self.app = app
        await self.initialize()
        await self.define_config(self.app.context.config)

    async def initialize(self):
        '''
        Method used to initialize the service.
        '''
        pass

    async def define_config(self, config):
        '''
        Method used to define extra configuration to the service.
        '''
        pass

    async def on_message(self, message):
        pass

    async def publish_message(self, socket_id, message_type, payload):
        p = self.app.connected_players.get(socket_id, None)
        if p is None:
            return

        msg = {
            'type': message_type,
            'socket_id': socket_id,
            'payload': payload,
        }
        await p.send_to_socket(dumps(msg))
