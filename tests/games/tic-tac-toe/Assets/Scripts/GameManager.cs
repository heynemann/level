using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;

public class GameManager : MonoBehaviour {
	public GameObject Server;
	public Color XColor;
	public Color OColor;

	public GameObject button0;
	public GameObject button1;
	public GameObject button2;
	public GameObject button3;
	public GameObject button4;
	public GameObject button5;
	public GameObject button6;
	public GameObject button7;
	public GameObject button8;

	public GameObject EndgamePanel;
	public GameObject WinText;
	public GameObject LostText;
	public GameObject DrawText;

	Text[] texts;

	T3Server serverComponent;
	string gameId;
	string playerSide;
	bool vsBot;
	Match game;

	void Awake() {
		Debug.Log ("AWAKE");
		gameId = PlayerPrefs.GetString ("currentMatchId");
		playerSide = PlayerPrefs.GetString ("playerSide");
		vsBot = PlayerPrefs.GetInt ("vsBot") == 1;
		serverComponent = Server.GetComponent<T3Server> ();
		EndgamePanel.SetActive (false);

		texts = new Text[] {
			button0.GetComponentInChildren<Text>(),
			button1.GetComponentInChildren<Text>(),
			button2.GetComponentInChildren<Text>(),
			button3.GetComponentInChildren<Text>(),
			button4.GetComponentInChildren<Text>(),
			button5.GetComponentInChildren<Text>(),
			button6.GetComponentInChildren<Text>(),
			button7.GetComponentInChildren<Text>(),
			button8.GetComponentInChildren<Text>(),
		};
	}
		
	void Start () {
		Debug.Log ("Start");
		EndgamePanel.SetActive (false);
		WinText.SetActive (false);
		LostText.SetActive (false);
		DrawText.SetActive (false);
		game = serverComponent.StartGame (gameId, playerSide, vsBot, OnStatusChange, OnMatchEnded);
	}

	void OnStatusChange(string matchId, string[] board) {
		UpdateBoard (board);
	}

	void UpdateBoard(string[] board) {
		for (var i = 0; i < board.Length; i++) {
			if (board [i] == "") {
				texts [i].text = "";
				continue;
			}
			if (board [i] == "x") {
				texts [i].text = "X";
				texts [i].color = XColor;
			}

			if (board [i] == "o") {
				texts [i].text = "O";
				texts [i].color = OColor;
			}
		}
	}

	void OnMatchEnded(string matchId, string[] board, string winner) {
		UpdateBoard (board);
		if (winner == "") {
			DrawText.SetActive (true);
		} else {
			WinText.SetActive (winner == playerSide);
			LostText.SetActive (winner != playerSide);
		}
		EndgamePanel.SetActive (true);
	}

	public void TakeAction(int pos) {
		game.TakeAction (pos);
	}

	public void GoHome() {
		SceneManager.LoadScene ("Main");
	}
}
