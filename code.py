def proof(trie, key):
  proofdb = new DB()
  node = trie.root
  while len(key) != 0:
    proofdb.put(node.hash, node)
    c = key[0]
    key = key[1:]
    node = db.get(node.branch[c])
    if node is None:
      return None
  proofdb.put(node.hash, node)
  return proofdb


def verifyProof(rootHash, key, proofdb):
  targetHash = rootHash
  i = 0
  while True:
    node = proofdb.get(targetHash)
    if node is None:
      return False, None
    if i == len(key):
      return True, node.value
    c = key[i]
    i += 1
    targetHash = node.branch[c]

