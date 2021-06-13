/*
 *  Synopsis:
 *	Find known unknowns.
 */
SET search_path TO secedgar,public;

SELECT
	e.blob
  FROM
  	edgar_put_daily e
	  LEFT OUTER JOIN nc_tar_file_element nc ON (
	  	nc.blob = e.blob
	  )
  WHERE
  	e.blob is null
;
