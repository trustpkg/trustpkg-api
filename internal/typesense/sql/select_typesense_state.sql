SELECT phase, watchlist_offset, last_package_id, cycle,
       documents_uploaded, watchlist_documents_uploaded,
       database_documents_uploaded
FROM typesense_state
WHERE id = 1;
