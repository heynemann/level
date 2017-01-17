using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using Facebook.Unity;
using UnityEngine.SceneManagement;

public class LoginManager : MonoBehaviour {
	void Start() {
		if (PlayerPrefs.GetString ("userId") != "") {
			GoToMain ();
		}
		FB.Init ();
	}

	void GoToMain() {
		SceneManager.LoadScene ("Main");
	}

	public void Login() {
		var perms = new List<string>(){"public_profile", "email", "user_friends"};
		FB.LogInWithReadPermissions(perms, AuthCallback);
	}
		
	private void AuthCallback (ILoginResult result) {
		if (FB.IsLoggedIn) {
			// AccessToken class will have session details
			var aToken = Facebook.Unity.AccessToken.CurrentAccessToken;

			// Print current access token's granted permissions
			foreach (string perm in aToken.Permissions) {
				Debug.Log(perm);
			}

			PlayerPrefs.SetString ("userId", aToken.UserId);
			GoToMain ();
		} else {
			Debug.Log("User cancelled login");
		}
	}
}
