using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;

public class MainManager : MonoBehaviour {
	public GameObject Server;
	public GameObject Matches;
	public GameObject Wins;
	public GameObject Ranking;

	T3Server serverComponent;

	// Use this for initialization
	void Start () {
		serverComponent = Server.GetComponent<T3Server> ();
		LoadUserDetails ();	
	}
	
	void LoadUserDetails() {
		var user = serverComponent.LoadUserDetails ();
		Matches.GetComponent<Text> ().text = user.Matches.ToString ();
		Wins.GetComponent<Text> ().text = user.Wins.ToString ();
		Ranking.GetComponent<Text> ().text = "#" + user.Ranking.ToString ();
	}

	public void EnterMatchmaking() {
		SceneManager.LoadScene ("MatchMaking");
	}
}
