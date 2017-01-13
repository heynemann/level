# level
[![Build Status](https://travis-ci.org/heynemann/level.svg?branch=master)](https://travis-ci.org/heynemann/level)
[![Coverage Status](https://coveralls.io/repos/github/heynemann/level/badge.svg?branch=master)](https://coveralls.io/github/heynemann/level?branch=master)

Level is a message-oriented game server.

## Features

* [x] HTTP and WebSocket connections powered by the insanely fast [Tornado](http://www.tornadoweb.org/en/stable/) web framework with the underlying networkings done by Python's [asyncio](https://docs.python.org/3/library/asyncio.html);
* [x] Modular architecture based in extensions and services;
* [x] 100% Horizontally scalable and cloud friendly;
* [ ] Built-in extensions for data storage ([MongoDB](https://www.mongodb.come)), PubSub ([Redis](https://redis.io/)), Caching ([Redis](https://redis.io/));
* [ ] Easily create custom extensions and plug them in;
* [ ] Social authentication based admin where you can manage your game server;
* [ ] Player Registration and Single Sign-On (for multiple devices);
* [ ] Virtual currency management and player currencies tracking;
* [ ] Game items management (CRUD and statistics);
* [ ] Game items purchasing - don't ever write code to verify if each player can buy each item again;
* [ ] Player inventory with unlocked items and item count;
* [ ] Offer system allowing players to purchase packs of items and currency (once, multiple times and/or time limited);
* [ ] Customizable party-based matchaking;
* [ ] Customizable room-based matchmaking;
* [ ] Game Loop service base - easily implement your own game loops without worrying about the networking;
* [ ] Clan service allowing players to participate in clans with little to no configuration using the battle-tested [Khan](https://github.com/topfreegames/khan);
* [ ] Leaderboard service enabling your players to compete in local and global, persistent or weekly rankings using [Podium](https://github.com/topfreegames/podium);
* [ ] Efficient logging allows for easy debugging;
* [ ] Level publishes statistics in whatever stats endpoint you want (built-in [statsd](https://github.com/etsy/statsd) and [elasticsearch](https://www.elastic.co/));
* [ ] 100% Code Coverage - be confident that you are using a safe server to handle your game;
* [ ] Comprehensive documentation detailing how to use, set-up and extend Level;
* [x] Level is a Python >= 3.6.0 library, meaning we won't support older versions of python (please don't ask);
* [x] MIT License ensures you won't be locked in or have to pay royalties later on;
* [ ] Client libraries for iOS (Objective-C), Android (Java) and Unity (C#).

## Getting Started

TBW.

## Contributions

"But I want to help!".

Great, just make sure you do the following:

* Fork;
* Get to a branch with the name of the feature you are implementing;
* Implement said feature (with tests and documentation changes);
* Send Pull Request;
* Rinse and repeat.

## License

Level and all its constituent services are MIT-Licensed unless expressed otherwise.
