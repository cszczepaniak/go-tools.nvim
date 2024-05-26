local M = {}

function M.run()
	if vim.bo.filetype ~= "go" then
		return
	end

	local pos = vim.api.nvim_win_get_cursor(0)
	local line = pos[1]
	local col = pos[2] + 1

	local file = vim.fn.expand("%")

	local res = vim.system({
		"go-tools",
		file .. "," .. tostring(line) .. "," .. tostring(col),
	}, { text = true }):wait()

	if res.code ~= 0 then
		vim.notify("go-tools exited with non-zero code\n" .. "stdout:\n" .. res.stderr, vim.log.levels.ERROR, {})
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
