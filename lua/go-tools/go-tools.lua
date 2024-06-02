local M = {}

function M.run()
	if vim.bo.filetype ~= "go" then
		return
	end

	-- cursor_bytes is 1-indexed, but the Go side will want it to be 0-indexed.
	local pos = vim.fn.wordcount().cursor_bytes - 1
	local file = vim.fn.expand("%")

	local res = vim.system({
		"go-tools",
		file .. "," .. tostring(pos),
	}, {
		text = true,
		stdin = vim.api.nvim_buf_get_lines(0, 0, -1, false),
	}):wait()

	if res.code ~= 0 then
		vim.notify(
			"go-tools exited with non-zero code\n" .. "stdout:\n" .. res.stdout .. "stderr:\n" .. res.stderr,
			vim.log.levels.ERROR,
			{}
		)
		return
	end

	if res.stdout == "" then
		return
	end

	local output = vim.json.decode(res.stdout)

	vim.api.nvim_buf_set_text(
		0,
		output.rng.start.ln - 1,
		output.rng.start.col - 1,
		output.rng.stop.ln - 1,
		output.rng.stop.col - 1,
		output.lns
	)
end

return M
