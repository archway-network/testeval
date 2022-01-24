package report

import "fmt"

func getHTMLHeader(title string, homePage string) string {

	return fmt.Sprintf(`<!doctype html>
	<html lang="en">
	  <head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		<title>Testnet Evaluator | %s</title>
		<!-- Bootstrap core CSS -->
		<link href="https://getbootstrap.com/docs/4.0/dist/css/bootstrap.min.css" rel="stylesheet">
	  </head>
	<body>
	<div class="d-flex flex-column flex-md-row align-items-center p-3 px-md-4 mb-3 bg-white border-bottom box-shadow">
      <h5 class="my-0 mr-md-auto font-weight-normal">%s</h5>
      <nav class="my-2 my-md-0 mr-md-3">
        <!-- <a class="p-2 text-dark" href="#">SomeLink</a> -->
      </nav>
      <a class="btn btn-outline-primary" href="%s">Home</a>
    </div>
	`, title, title, homePage)
}

func getHTMLFooter() string {
	return `</body></html>`
}

func getHTMLTable(headers []string, rows [][]string, footers []string) string {

	out := `<div class="table-responsive">
	<table class="table table-striped table-hover">
		<thead class="thead-dark">
			<tr><th scope="col">#</th>`

	for i := range headers {
		out += fmt.Sprintf(`<th scope="col">%s</th>`, headers[i])
	}
	out += `<tbody>`

	for i := range rows {
		out += fmt.Sprintf(`<tr><th scope="row">%d</th>`, i+1)

		for j := range rows[i] {
			out += fmt.Sprintf(`<td>%s</td>`, rows[i][j])
		}
		out += `</tr>`
	}

	if footers != nil {
		out += `<tr class="table-info"><td scope="col"></td>`
		for i := range footers {
			out += fmt.Sprintf(`<th scope="col">%s</th>`, footers[i])
		}
		out += `</tr>`
	}

	out += `</tbody></table></div>`
	return out
}
