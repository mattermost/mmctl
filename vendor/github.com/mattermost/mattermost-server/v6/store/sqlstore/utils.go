// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"
	"strconv"
	"strings"
	"unicode"

	"github.com/mattermost/gorp"

	"github.com/mattermost/mattermost-server/v6/shared/mlog"
)

var escapeLikeSearchChar = []string{
	"%",
	"_",
}

func sanitizeSearchTerm(term string, escapeChar string) string {
	term = strings.Replace(term, escapeChar, "", -1)

	for _, c := range escapeLikeSearchChar {
		term = strings.Replace(term, c, escapeChar+c, -1)
	}

	return term
}

// Converts a list of strings into a list of query parameters and a named parameter map that can
// be used as part of a SQL query.
func MapStringsToQueryParams(list []string, paramPrefix string) (string, map[string]interface{}) {
	var keys strings.Builder
	params := make(map[string]interface{}, len(list))
	for i, entry := range list {
		if keys.Len() > 0 {
			keys.WriteString(",")
		}

		key := paramPrefix + strconv.Itoa(i)
		keys.WriteString(":" + key)
		params[key] = entry
	}

	return "(" + keys.String() + ")", params
}

// finalizeTransaction ensures a transaction is closed after use, rolling back if not already committed.
func finalizeTransaction(transaction *gorp.Transaction) {
	// Rollback returns sql.ErrTxDone if the transaction was already closed.
	if err := transaction.Rollback(); err != nil && err != sql.ErrTxDone {
		mlog.Error("Failed to rollback transaction", mlog.Err(err))
	}
}

// finalizeTransactionX ensures a transaction is closed after use, rolling back if not already committed.
func finalizeTransactionX(transaction *sqlxTxWrapper) {
	// Rollback returns sql.ErrTxDone if the transaction was already closed.
	if err := transaction.Rollback(); err != nil && err != sql.ErrTxDone {
		mlog.Error("Failed to rollback transaction", mlog.Err(err))
	}
}

// removeNonAlphaNumericUnquotedTerms removes all unquoted words that only contain
// non-alphanumeric chars from given line
func removeNonAlphaNumericUnquotedTerms(line, separator string) string {
	words := strings.Split(line, separator)
	filteredResult := make([]string, 0, len(words))

	for _, w := range words {
		if isQuotedWord(w) || containsAlphaNumericChar(w) {
			filteredResult = append(filteredResult, strings.TrimSpace(w))
		}
	}
	return strings.Join(filteredResult, separator)
}

// containsAlphaNumericChar returns true in case any letter or digit is present, false otherwise
func containsAlphaNumericChar(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// isQuotedWord return true if the input string is quoted, false otherwise. Ex :-
// 		"quoted string"  -  will return true
// 		unquoted string  -  will return false
func isQuotedWord(s string) bool {
	if len(s) < 2 {
		return false
	}

	return s[0] == '"' && s[len(s)-1] == '"'
}

// constructMySQLJSONArgs returns the arg list to pass to a query along with
// the string of placeholders which is needed to be to the JSON_SET function.
// Use this function in this way:
// UPDATE Table
// SET Col = JSON_SET(Col, `+argString+`)
// WHERE Id=?`, args...)
// after appending the Id param to the args slice.
func constructMySQLJSONArgs(props map[string]string) ([]interface{}, string) {
	if len(props) == 0 {
		return nil, ""
	}

	// Unpack the keys and values to pass to MySQL.
	args := make([]interface{}, 0, len(props))
	for k, v := range props {
		args = append(args, "$."+k)
		args = append(args, v)
	}

	// We calculate the number of ? to set in the query string.
	argString := strings.Repeat("?, ", len(props)*2)
	// Strip off the trailing comma.
	argString = strings.TrimSuffix(strings.TrimSpace(argString), ",")

	return args, argString
}

func makeStringArgs(params []string) []interface{} {
	args := make([]interface{}, len(params))
	for i, name := range params {
		args[i] = name
	}
	return args
}

func constructArrayArgs(ids []string) (string, []interface{}) {
	var placeholder strings.Builder
	values := make([]interface{}, 0, len(ids))
	for _, entry := range ids {
		if placeholder.Len() > 0 {
			placeholder.WriteString(",")
		}

		placeholder.WriteString("?")
		values = append(values, entry)
	}

	return "(" + placeholder.String() + ")", values
}
