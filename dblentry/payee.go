package dblentry

import "regexp"

type Payee struct {
	name      string
	aliases   []string // regular expression that can match transactions.
	realiases []*regexp.Regexp
	uuids     []string
}

func NewPayee(name string) *Payee {
	return &Payee{
		name:      name,
		aliases:   []string{},
		realiases: []*regexp.Regexp{},
		uuids:     []string{},
	}
}

func (payee *Payee) addAlias(pattern string) error {
	if pattern == "" {
		return nil
	}
	payee.aliases = append(payee.aliases, pattern)
	regc, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	payee.realiases = append(payee.realiases, regc)
	return nil
}

func (payee *Payee) addUuid(uuid string) {
	if uuid == "" {
		return
	}
	payee.uuids = append(payee.uuids, uuid)
}

func (payee *Payee) matchAlias(alias string) (string, bool) {
	for _, regc := range payee.realiases {
		if regc.MatchString(alias) {
			return payee.name, true
		}
	}
	return "", false
}

func (payee *Payee) matchUuid(uuid string) (string, bool) {
	for _, xuuid := range payee.uuids {
		if xuuid == uuid {
			return payee.name, true
		}
	}
	return "", false
}
