# HR Sistem  
Ovaj projekat je razvijen u okviru predmeta **WEB2** na **Prirodno-matematičkom fakultetu u Kragujevcu**. 
Cilj projekta je implementacija centralizovanog HR sistema sa fokusom na jasnu arhitekturu, upravljanje korisnicima i praćenje poslovnih procesa.  

## Opis projekta  
Aplikacija omogućava upravljanje zaposlenima, njihovim odsustvima i evidencijom prisustva, uz praćenje svih promena u sistemu kroz audit log. Sistem je zasnovan na role-based pristupu, gde različiti tipovi korisnika (admin, menadžer, zaposleni) imaju definisana prava pristupa i akcije koje mogu izvršavati.  Glavne funkcionalnosti uključuju: 
- autentifikaciju i autorizaciju korisnika (JWT) 
- upravljanje zaposlenima (pregled, kreiranje, izmena) 
- podnošenje i odobravanje zahteva za odsustvo 
- evidenciju dolaska i odlaska zaposlenih 
- audit log svih izmena u sistemu  

## Tehnologije  
- **Backend:** Go (Golang), REST API 
- **Frontend:** React 
- **Baze podataka:**   
  - PostgreSQL (primarna baza)   
  - MongoDB (audit log i istorijski podaci)

## Arhitektura  
Backend je organizovan slojevito: 
- handler (HTTP sloj)
- service (poslovna logika)
- repository (pristup bazi)  

Ovakav pristup omogućava bolju testabilnost, održavanje i proširivost sistema.  

## Pokretanje projekta  
Aplikacija je predviđena za pokretanje u lokalnom okruženju korišćenjem kontejnera (Podman/Docker), uz konfiguraciju putem `.env` fajlova.
