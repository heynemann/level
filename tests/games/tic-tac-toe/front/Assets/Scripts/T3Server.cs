using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using Facebook.Unity;
using UnityEngine.SceneManagement;
using System;

public delegate void OnStatusChangeDelegate(string matchId, string[] board);
public delegate void OnMatchEndedDelegate(string matchId, string[] board, string winner);

public class User {
	int _matches, _wins, _ranking;

	public User(int matches, int wins, int ranking) {
		_matches = matches;
		_wins = wins;
		_ranking = ranking;
	}

	public int Matches {
		get { return _matches; }
	}

	public int Wins {
		get { return _wins; }
	}

	public int Ranking {
		get { return _ranking; }
	}
}

public class Match {
	string _matchId, _playerSide;
	bool _vsBot, _gameOver;
	OnStatusChangeDelegate _onStatusChange;
	OnMatchEndedDelegate _onMatchEnded;
	string[] board;

	public Match(string matchId, string playerSide, bool vsBot, OnStatusChangeDelegate onStatusChange, OnMatchEndedDelegate onMatchEnded) {
		_matchId = matchId;
		_playerSide = playerSide;
		_vsBot = vsBot;
		_onStatusChange = onStatusChange;
		_onMatchEnded = onMatchEnded;
		_gameOver = false;
		board = new string[]{ "", "", "", "", "", "", "", "", "" };
	}

	public void Start() {
		// Right now we do nothing. Later on we'll connect to the server.
	}

	public void TakeAction(int pos) {
		if (_gameOver || IsTaken (pos)) {
			return;
		}

		board [pos] = _playerSide;

		if (GameEnded ()) {
			GameOver ();
			return;
		}

		if (_vsBot) {
			var nextPos = getNextBotPos ();
			if (nextPos == -1) {
				GameOver ();
				return;
			}
			board [nextPos] = _playerSide == "x" ? "o" : "x"; 
		}

		if (GameEnded ()) {
			GameOver ();
			return;
		}

		_onStatusChange (_matchId, board);
	}

	bool IsTaken(int pos) {
		return board [pos] != "";
	}

	int AvailableMoves() {
		int moves = 0;
		for (var i = 0; i < board.Length; i++) {
			if (board [i] == "") {
				moves++;
			}
		}
		return moves;
	}

	bool GameEnded() {
		if (AvailableMoves() == 0) {
			return true;
		}
		if (PlayerWon ("x") || PlayerWon ("o")) {
			return true;
		}
		return false;
	}

	bool PlayerWon(string playerSide) {
		if (
			//Horizontal win
			(board [0] == playerSide && board [1] == playerSide && board [2] == playerSide) ||
			(board [3] == playerSide && board [4] == playerSide && board [5] == playerSide) ||
			(board [6] == playerSide && board [7] == playerSide && board [8] == playerSide) ||

			//Vertical win
			(board [0] == playerSide && board [3] == playerSide && board [6] == playerSide) ||
			(board [1] == playerSide && board [4] == playerSide && board [7] == playerSide) ||
			(board [2] == playerSide && board [5] == playerSide && board [8] == playerSide) ||

			//Diagonal win
			(board [0] == playerSide && board [4] == playerSide && board [8] == playerSide) ||
			(board [2] == playerSide && board [4] == playerSide && board [6] == playerSide)
			) {
			return true;
		}

		return false;
	}

	void GameOver() {
		_gameOver = true;
		var winner = PlayerWon ("x") ? "x" : PlayerWon ("o") ? "o" : "";
		_onMatchEnded (_matchId, board, winner);
	}

	int getNextBotPos() {
		for (var i = 0; i < board.Length; i++) {
			if (board [i] == "") {
				return i;
			}
		}

		return -1;
	}
}


public class T3Server : MonoBehaviour {
	public delegate void MatchFoundDelegate(string matchId, string playerSide, bool vsBot);

	public User LoadUserDetails() {
		return new User (
			100, 50, 124
		);
	}

	public void EnterMatchMaking(MatchFoundDelegate onMatchFound) {
		onMatchFound (Guid.NewGuid().ToString(), "x", true);
	}

	public Match StartGame(string gameId, string playerSide, bool vsBot, OnStatusChangeDelegate onStatusChange, OnMatchEndedDelegate onMatchEnded) {
		var game = new Match (gameId, playerSide, vsBot, onStatusChange, onMatchEnded);
		game.Start ();

		return game;
	}
}