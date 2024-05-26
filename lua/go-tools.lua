local M = {}

function M.init()
	local tools = require("go-tools.go-tools")
	vim.api.nvim_create_user_command("GoToolsOmni", tools.run, {})
	vim.keymap.set("n", "<leader>go", "<cmd>GoToolsOmni<CR>", { desc = "[G]o tools [o]mni function" })
end

return M
