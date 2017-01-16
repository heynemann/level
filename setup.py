#!/usr/bin/env python
# -*- coding: utf-8 -*-

# This file is part of level.
# https://github.com/heynemann/level

# Licensed under the MIT license:
# http://www.opensource.org/licenses/MIT-license
# Copyright (c) 2016, Bernardo Heynemann <heynemann@gmail.com>

from setuptools import setup, find_packages
from level import __version__

tests_require = [
    'mock',
    'nose',
    'coverage',
    'yanc',
    'preggy',
    'tox',
    'ipdb',
    'coveralls',
    'sphinx',
    'nose-focus',
]

setup(
    name='level',
    version=__version__,
    description='Level is a message-oriented gaming server.',
    long_description='''
Level is a message-oriented gaming server.
''',
    keywords='gaming game server service websocket',
    author='Bernardo Heynemann',
    author_email='heynemann@gmail.com',
    url='https://github.com/heynemann/level',
    license='MIT',
    classifiers=[
        'Development Status :: 4 - Beta',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: MIT License',
        'Natural Language :: English',
        'Operating System :: Unix',
        'Programming Language :: Python :: 3.4',
        'Operating System :: OS Independent',
    ],
    packages=find_packages(),
    include_package_data=True,
    install_requires=[
        # add your dependencies here
        # remember to use 'package-name>=x.y.z,<x.(y+1).0' notation
        # (this way you get bugfixes but no breaking changes)
        'tornado>=4.4.2,<5.0.0',
        'toredis-fork>=0.1.4,<1.0.0',
        'importer-lib>=0.2.0,<1.0.0',
        'cement>=2.10.2,<3.0.0',
        'derpconf>=0.8.1,<1.0.0',
        'ujson>=1.35,<2.0',
    ],
    extras_require={
        'tests': tests_require,
    },
    entry_points={
        'console_scripts': [
            # add cli scripts here in this form:
            'level=level.cmd.cli:main',
        ],
    },
)
