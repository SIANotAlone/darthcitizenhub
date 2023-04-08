get news json endpoints:
    /news/games
    /news/films
    /news/serials

get author news json endpoint:
    /news/darthcitizen

post author news json endpoint:
    Method POST
    JWT protected
    /admin/postnews

update author news json endpoint:
    Method PUT
    JWT protected
    /admin/update/{id}

delete author news json endpoint:
    Method DELETE
    JWT protected
    /admin/delete/{id}