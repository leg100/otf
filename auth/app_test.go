package auth

type fakeApp struct {
	*fakeTeamApp
	*fakeAgentTokenApp

	registrySessionApp
	sessionApp
	userApp
}
