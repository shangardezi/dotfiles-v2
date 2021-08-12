-- check Packer is installed
local execute = vim.api.nvim_command
local fn = vim.fn

local install_path = fn.stdpath('data')..'/site/pack/packer/start/packer.nvim'

if fn.empty(fn.glob(install_path)) > 0 then
  fn.system({'git', 'clone', 'https://github.com/wbthomason/packer.nvim', install_path})
  execute 'packadd packer.nvim'
end

vim.cmd [[packadd packer.nvim]]

return require('packer').startup(function(use)
  -- Packer can manage itself
  use 'wbthomason/packer.nvim'

  -- LSP
  use 'neovim/nvim-lspconfig' -- bootstrap LSP configuration
  use 'kabouzeid/nvim-lspinstall' -- install any LSP server
  use 'hrsh7th/nvim-compe' -- autocompletion
  use 'hrsh7th/vim-vsnip' -- LSP based snippets
  use 'hrsh7th/vim-vsnip-integ'
  --use 'golang/vscode-go' -- snippets like it's hot (enabled by vim-snip): manually imported the snippets I use with the prefixes I'm used to
  use 'onsails/lspkind-nvim' -- add pictograms to autocompletion LSP results
  use 'glepnir/lspsaga.nvim' -- some UI for LSP <- under trial
  use {
    'nvim-treesitter/nvim-treesitter',
    run = ':TSUpdate'
  } -- just the best thing
  use {'npxbr/gruvbox.nvim', requires = {'rktjmp/lush.nvim'}} -- sick theme
  use 'joshdick/onedark.vim' -- Atom's theme

  use 'jiangmiao/auto-pairs'
  use 'tpope/vim-sensible'
  use 'myusuf3/numbers.vim'
  use 'danishprakash/vim-githubinator'
  use 'tpope/vim-fugitive'
  use 'scrooloose/nerdcommenter'
  use 'tpope/vim-surround'
  use 'AndrewRadev/splitjoin.vim'

  use 'nvim-lua/popup.nvim'
  use 'nvim-lua/plenary.nvim'
  use 'nvim-telescope/telescope.nvim'
  use 'pwntester/octo.nvim'
  use 'kyazdani42/nvim-web-devicons'
  use 'kyazdani42/nvim-tree.lua'
end)
