package auth

import "time"

// Dve različite „dužine trajanja" u sistemu — najčešća tačka zabune:
//
//	accessTokenTTL (password.go) → koliko dugo važi JEDAN JWT (jedan zahtev)
//	SessionTTL     (ovaj fajl)   → koliko dugo traje CELA SESIJA
//
// Kada access token istekne, klijent ga tiho osveži preko refresh kolačića —
// zato 20 minuta ne znači „posle 20 minuta si izbačen", nego „posle 20 minuta
// mora da se osveži".
//
// Sesija traje najviše SessionTTL **od prijave** (apsolutno, ne produžava se
// osvežavanjem) i upisuje se u `refresh_tokens.expires_at`. Nezavisno od toga,
// refresh kolačić je „session cookie", pa ga pretraživač briše i pri zatvaranju.
// Sesija se, dakle, završava šta prvo nastupi: zatvaranje pretraživača ili istek
// SessionTTL-a.
const SessionTTL = 8 * time.Hour
