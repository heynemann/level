using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.SceneManagement;

public class MatchmakingManager : MonoBehaviour {
	public GameObject Server;
	T3Server serverComponent;

	// Use this for initialization
	void Start () {
		serverComponent = Server.GetComponent<T3Server> ();
		serverComponent.EnterMatchMaking(OnMatchFound);
	}

	void OnMatchFound(string matchId, string playerSide, bool vsBot) {
		PlayerPrefs.SetString ("currentMatchId", matchId);
		PlayerPrefs.SetInt ("vsBot", vsBot ? 1 : 0);
		PlayerPrefs.SetString ("playerSide", playerSide);

		SceneManager.LoadScene ("Game");
	}
}
