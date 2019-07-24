package types

type Account struct {
	ID                            ObjectID   `json:"id"`
	MembershipExpirationDate      Time       `json:"membership_expiration_date"`
	Registrar                     ObjectID   `json:"registrar"`
	Referrer                      ObjectID   `json:"referrer"`
	LifetimeReferrer              ObjectID   `json:"lifetime_referrer"`
	NetworkFeePercentage          uint32     `json:"network_fee_percentage"`
	LifetimeReferrerFeePercentage uint32     `json:"lifetime_referrer_fee_percentage"`
	ReferrerRewardsPercentage     uint32     `json:"referrer_rewards_percentage"`
	Name                          string     `json:"name"`
	Statistics                    string     `json:"statistics"`
	Options                       Options    `json:"options"`
	Owner                         Permission `json:"owner"`
	Active                        Permission `json:"active"`
}

type Permission struct {
	WeightThreshold uint32        `json:"weight_threshold"`
	AccountAuths    []interface{} `json:"account_auths"`
	KeyAuths        []interface{} `json:"key_auths"`
	AddressAuths    []interface{} `json:"address_auths"`
}

type Options struct {
	MemoKey       string   `json:"memo_key"`
	VotingAccount ObjectID `json:"voting_account"`
}
