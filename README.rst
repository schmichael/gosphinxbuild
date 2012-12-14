Sphinx Autobuilder
==================

So you write Sphinx_ docs? That's great. Want an autobuilder? Here's one:

.. code-block::

     go get -u github.com/schmichael/gosphinxbuild 
     cd ~/some/sphinx/root
     gsb

Just run ``gsb`` in an environment where ``make html`` works (so activate a
virtualenv if necessary).

.. _Sphinx: http://sphinx-doc.org/
