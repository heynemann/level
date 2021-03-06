#!/usr/bin/env python
# -*- coding: utf-8 -*-

# this file is part of level.
# https://github.com/heynemann/level

# licensed under the mit license:
# http://www.opensource.org/licenses/mit-license
# copyright (c) 2016, bernardo heynemann <heynemann@gmail.com>

import sys
import signal

from tornado.ioloop import IOLoop
from cement.core.exc import CaughtSignal
from cement.core.foundation import CementApp
from cement.core.controller import CementBaseController, expose

from level.context import ServerParameters
from level.server import run
from level.config import generate_config
from level.load import LoadTest


LOG_LEVELS = {
    0: 'error',
    1: 'warn',
    2: 'info',
    3: 'debug',
}


class LevelController(CementBaseController):
    class Meta:
        label = 'base'
        description = 'Use this command to manage your Level server.'


class LevelUpController(CementBaseController):
    class Meta:
        label = 'up'
        description = 'Start your Level server.'
        stacked_on = 'base'
        stacked_type = 'nested'

        arguments = [
            (['-c', '--config'], dict(help='Configuration file path.')),
            (['-p', '--port'], dict(default=8888, type=int, help='Port to bind Level server to.')),
            (['-b', '--bind'], dict(default='localhost', help='Host to bind Level server to.')),
            (['-v', '--verbose'], dict(default=0, action='count', help='Log level (0-error, 1-warn, 2-info, 3-debug).')),
        ]

    @expose(help='Starts a configured level server.')
    def default(self):
        params = ServerParameters(
            io_loop=IOLoop.instance(),
            host=self.app.pargs.bind,
            port=int(self.app.pargs.port),
            config_path=self.app.pargs.config,
            log_level=LOG_LEVELS[self.app.pargs.verbose],
            debug=self.app.pargs.debug,
        )

        try:
            run(params)
        except CaughtSignal as e:
            if e.signum == signal.SIGTERM:
                sys.stdout.write('\n')
                sys.stdout.write("-- level closed by SIGTERM --\n")
            elif e.signum == signal.SIGINT:
                sys.stdout.write("-- level closed by SIGINT --\n")
            else:
                sys.stdout.write("-- level closed by Signal: %d --\n" % e.signum)


class LevelConfigController(CementBaseController):
    class Meta:
        label = 'config'
        stacked_on = 'base'
        description = 'Generate a base configuration file.'
        stacked_type = 'nested'

    @expose(help='Generates a base configuration file.')
    def default(self):
        generate_config()


class LevelLoadController(CementBaseController):
    class Meta:
        label = 'load'
        stacked_on = 'base'
        description = 'Load test level\'s websocket infrastructure.'
        stacked_type = 'nested'

        arguments = [
            (['-c', '--concurrency'], dict(type=int, default=200, help='Amount of concurrent websocket connections.')),
            (['-d', '--duration'], dict(type=int, default=10, help='Number of seconds to run test with.')),
            (['-r', '--recycle'], dict(type=int, default=2, help='Number of seconds to keep websocket connection alive.')),

        ]

    @expose(help='Load test level\'s websocket infrastructure.')
    def default(self):
        load = LoadTest(
            'ws://localhost:8888/ws',
            int(self.app.pargs.concurrency),
            int(self.app.pargs.recycle),
            int(self.app.pargs.duration),
        )
        try:
            load.start()
        except Exception:
            load.stop()


class LevelCliApp(CementApp):
    class Meta:
        label = 'level'
        framework_logging = False
        handlers = [
            LevelController,
            LevelUpController,
            LevelConfigController,
            LevelLoadController,
        ]


def main():
    with LevelCliApp() as app:
        app.run()


if __name__ == "__main__":
    main()
