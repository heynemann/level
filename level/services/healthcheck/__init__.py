#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>


from level.services import BaseService
from level.handlers import BaseHandler


class HealthcheckHandler(BaseHandler):
    async def do_healthcheck(self):
        return {
            'success': True,
        }

    async def get(self):
        result = await self.do_healthcheck()
        self.json(result)

    async def head(self, *args, **kwargs):
        result = await self.do_healthcheck()
        if result['success']:
            status = 200
        else:
            status = 500

        self.set_status(status)


class Service(BaseService):
    async def get_handlers(self):
        return (
            ('/healthcheck', HealthcheckHandler),
        )
